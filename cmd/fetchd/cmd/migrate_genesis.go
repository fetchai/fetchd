package cmd

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

// MigrateGenesisCmd migrates a legacy (e.g., v0.47) genesis.json so it loads under SDK v0.53.
// It can also initialize missing modules using ModuleBasics.DefaultGenesis.
func MigrateGenesisCmd(basicManager module.BasicManager) *cobra.Command {
	var (
		stripModsCSV string
		initMissing  bool
		initOnlyCSV  string // if set, only initialize these modules (subset)
	)

	cmd := &cobra.Command{
		Use:   "migrate-genesis [old_genesis.json] [new_genesis.json]",
		Short: "Migrate a legacy genesis.json for Cosmos SDK v0.53; can also init missing modules",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inFile := args[0]
			outFile := args[1]

			// Load source genesis
			appState, appGen, err := genutiltypes.GenesisStateFromGenFile(inFile)
			if err != nil {
				return fmt.Errorf("read genesis %q: %w", inFile, err)
			}

			// Strip legacy/removed modules (if requested)
			for _, m := range parseCSV(stripModsCSV) {
				if _, ok := appState[m]; ok {
					delete(appState, m)
					fmt.Fprintf(cmd.ErrOrStderr(), "stripped module: %s\n", m)
				}
			}

			// Initialize missing modules with defaults (optional)
			if initMissing {
				ir := codectypes.NewInterfaceRegistry()
				basicManager.RegisterInterfaces(ir)
				cdc := codec.NewProtoCodec(ir)

				defaults := basicManager.DefaultGenesis(cdc)
				only := toStringSet(parseCSV(initOnlyCSV))

				for mod, def := range defaults {
					if _, exists := appState[mod]; exists {
						continue
					}
					if len(only) > 0 {
						if _, ok := only[mod]; !ok {
							continue
						}
					}
					appState[mod] = def
					fmt.Fprintf(cmd.OutOrStdout(), "initialized missing module with default genesis: %s\n", mod)
				}
			}

			// Core sanitizers / migrations (chain-agnostic + your custom bits)
			if err := resetGovPreserveParamsAndFixBank(appState); err != nil {
				return err
			}

			if err := sanitizeMint(appState); err != nil {
				return err
			}

			genTime, initHeight, err := readGenesisHeader(inFile)
			if err != nil {
				return err
			}

			if err := ensureMintParamsSafe(appState, "afet"); err != nil {
				return err
			}
			if err := sanitizeFeegrant(appState, genTime); err != nil {
				return err
			}
			if err := migrateIBCTransfer(appState); err != nil {
				return err
			}
			if err := migrateWasm(appState, initHeight); err != nil {
				return err
			}

			// Write back
			newAppState, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("marshal new app_state: %w", err)
			}
			appGen.AppState = newAppState

			if err := ensureDir(outFile); err != nil {
				return err
			}
			if err := genutil.ExportGenesisFile(appGen, outFile); err != nil {
				return fmt.Errorf("write new genesis %q: %w", outFile, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Migrated genesis written to %s\n", outFile)
			return nil
		},
	}

	// Flags
	cmd.Flags().StringVar(&stripModsCSV, "strip-modules", "capability,crisis",
		"Comma-separated module names to remove from app_state (legacy/unused)")
	cmd.Flags().BoolVar(&initMissing, "init-missing-modules", true,
		"If true, initialize any missing modules using ModuleBasics.DefaultGenesis")
	cmd.Flags().StringVar(&initOnlyCSV, "init-only", "",
		"Optional comma-separated allowlist of module names to initialize (implies --init-missing-modules)")

	return cmd
}

// readGenesisHeader reads genesis_time and initial_height from a genesis JSON file.
func readGenesisHeader(path string) (time.Time, uint64, error) {
	var hdr struct {
		GenesisTime   string `json:"genesis_time"`
		InitialHeight string `json:"initial_height"`
	}
	bz, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("read genesis: %w", err)
	}
	if err := json.Unmarshal(bz, &hdr); err != nil {
		return time.Time{}, 0, fmt.Errorf("unmarshal header: %w", err)
	}

	// Parse time
	var t time.Time
	if hdr.GenesisTime == "" {
		t = time.Unix(0, 0).UTC()
	} else if tt, err := time.Parse(time.RFC3339Nano, hdr.GenesisTime); err == nil {
		t = tt.UTC()
	} else if tt, err := time.Parse(time.RFC3339, hdr.GenesisTime); err == nil {
		t = tt.UTC()
	} else {
		return time.Time{}, 0, fmt.Errorf("parse genesis_time: %w", err)
	}

	// Parse height
	var h uint64
	if hdr.InitialHeight != "" {
		u, err := strconv.ParseUint(hdr.InitialHeight, 10, 64)
		if err != nil {
			return time.Time{}, 0, fmt.Errorf("parse initial_height: %w", err)
		}
		h = u
	}
	return t, h, nil
}

// Reset gov to a minimal v1 skeleton, preserve params, and reconcile bank:
// - preserve gov.params (or legacy split params) + starting_proposal_id
// - clear proposals/deposits/votes
// - zero gov module account balance and subtract from bank supply
func resetGovPreserveParamsAndFixBank(appState map[string]json.RawMessage) error {
	type coin struct{ Denom, Amount string }

	// Extract params from old gov (v1 or v1beta1)
	var (
		startingID = "1"
		paramsOut  = map[string]any{}
	)
	if rawGov, ok := appState["gov"]; ok && len(rawGov) > 0 {
		var old map[string]any
		if err := json.Unmarshal(rawGov, &old); err == nil {
			if v, ok := old["starting_proposal_id"].(string); ok && v != "" {
				startingID = v
			}
			// v1 params
			if p, ok := old["params"].(map[string]any); ok {
				for k, v := range p {
					paramsOut[k] = v
				}
			}
			// legacy split params
			if dp, ok := old["deposit_params"].(map[string]any); ok {
				if v, ok := dp["min_deposit"]; ok {
					paramsOut["min_deposit"] = v
				}
				if v, ok := dp["max_deposit_period"]; ok {
					paramsOut["max_deposit_period"] = v
				}
			}
			if vp, ok := old["voting_params"].(map[string]any); ok {
				if v, ok := vp["voting_period"]; ok {
					paramsOut["voting_period"] = v
				}
			}
			if tp, ok := old["tally_params"].(map[string]any); ok {
				if v, ok := tp["quorum"]; ok {
					paramsOut["quorum"] = v
				}
				if v, ok := tp["threshold"]; ok {
					paramsOut["threshold"] = v
				}
				if v, ok := tp["veto_threshold"]; ok {
					paramsOut["veto_threshold"] = v
				}
			}
		}
	}

	// Ensure min_deposit exists (derive from staking bond_denom or bank supply)
	ensureMinDeposit := func() {
		if _, ok := paramsOut["min_deposit"]; ok {
			return
		}
		if rawStaking, ok := appState["staking"]; ok && len(rawStaking) > 0 {
			var st map[string]any
			if json.Unmarshal(rawStaking, &st) == nil {
				if p, ok := st["params"].(map[string]any); ok {
					if den, ok := p["bond_denom"].(string); ok && den != "" {
						paramsOut["min_deposit"] = []map[string]string{{"denom": den, "amount": "0"}}
						return
					}
				}
			}
		}
		if rawBank, ok := appState["bank"]; ok && len(rawBank) > 0 {
			var bk map[string]any
			if json.Unmarshal(rawBank, &bk) == nil {
				if sup, ok := bk["supply"].([]any); ok && len(sup) > 0 {
					if c, ok := sup[0].(map[string]any); ok {
						if den, _ := c["denom"].(string); den != "" {
							paramsOut["min_deposit"] = []map[string]string{{"denom": den, "amount": "0"}}
							return
						}
					}
				}
			}
		}
		paramsOut["min_deposit"] = []any{}
	}
	ensureMinDeposit()

	// Fill optional v1 params if absent
	if _, ok := paramsOut["min_initial_deposit_ratio"]; !ok {
		paramsOut["min_initial_deposit_ratio"] = "0.000000000000000000"
	}
	if _, ok := paramsOut["burn_vote_quorum"]; !ok {
		paramsOut["burn_vote_quorum"] = false
	}
	if _, ok := paramsOut["burn_proposal_deposit_prevote"]; !ok {
		paramsOut["burn_proposal_deposit_prevote"] = false
	}
	if _, ok := paramsOut["burn_vote_veto"]; !ok {
		paramsOut["burn_vote_veto"] = true
	}

	// Minimal gov v1 state
	minGov := map[string]any{
		"starting_proposal_id": startingID,
		"deposits":             []any{},
		"votes":                []any{},
		"proposals":            []any{},
		"params":               paramsOut,
	}
	if bz, err := json.Marshal(minGov); err == nil {
		appState["gov"] = bz
	} else {
		return fmt.Errorf("marshal new gov: %w", err)
	}

	// Reconcile bank: zero gov module account and subtract from supply
	govModAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	var bank map[string]any
	if rawBank, ok := appState["bank"]; ok && len(rawBank) > 0 {
		if err := json.Unmarshal(rawBank, &bank); err != nil {
			return fmt.Errorf("unmarshal bank: %w", err)
		}
	} else {
		return nil
	}

	parseCoins := func(v any) ([]struct{ Denom, Amount string }, bool) {
		arr, ok := v.([]any)
		if !ok {
			return nil, false
		}
		out := make([]struct{ Denom, Amount string }, 0, len(arr))
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				den, _ := m["denom"].(string)
				amt, _ := m["amount"].(string)
				if den != "" && amt != "" {
					out = append(out, struct{ Denom, Amount string }{Denom: den, Amount: amt})
				}
			}
		}
		return out, true
	}
	toAnyCoins := func(cs []struct{ Denom, Amount string }) []any {
		out := make([]any, len(cs))
		for i, c := range cs {
			out[i] = map[string]any{"denom": c.Denom, "amount": c.Amount}
		}
		return out
	}

	// Zero gov balance
	var balances []any
	if v, ok := bank["balances"]; ok {
		if arr, ok := v.([]any); ok {
			balances = arr
		}
	}
	removed := []struct{ Denom, Amount string }{}
	newBalances := make([]any, 0, len(balances))
	for _, b := range balances {
		m, ok := b.(map[string]any)
		if !ok {
			newBalances = append(newBalances, b)
			continue
		}
		addr, _ := m["address"].(string)
		if addr != govModAddr {
			newBalances = append(newBalances, b)
			continue
		}
		if cs, ok := parseCoins(m["coins"]); ok && len(cs) > 0 {
			removed = cs
		}
		m["coins"] = []any{} // keep entry, but zero coins
		newBalances = append(newBalances, m)
	}
	bank["balances"] = newBalances

	// Subtract from supply
	var supply []struct{ Denom, Amount string }
	if v, ok := bank["supply"]; ok {
		if cs, ok := parseCoins(v); ok {
			supply = cs
		}
	}
	if len(removed) > 0 && len(supply) > 0 {
		cur := map[string]*big.Int{}
		for _, c := range supply {
			v := new(big.Int)
			if _, ok := v.SetString(c.Amount, 10); !ok {
				v.SetInt64(0)
			}
			cur[c.Denom] = v
		}
		for _, r := range removed {
			if v, ok := cur[r.Denom]; ok {
				rm := new(big.Int)
				if _, ok := rm.SetString(r.Amount, 10); ok {
					v.Sub(v, rm)
					if v.Sign() < 0 {
						v.SetInt64(0)
					}
				}
			}
		}
		newSupply := make([]struct{ Denom, Amount string }, 0, len(cur))
		for den, v := range cur {
			newSupply = append(newSupply, struct{ Denom, Amount string }{Denom: den, Amount: v.String()})
		}
		sort.Slice(newSupply, func(i, j int) bool { return newSupply[i].Denom < newSupply[j].Denom })
		bank["supply"] = toAnyCoins(newSupply)
	}

	nb, err := json.Marshal(bank)
	if err != nil {
		return fmt.Errorf("marshal bank: %w", err)
	}
	appState["bank"] = nb
	return nil
}

// sanitizeMint drops unknown custom fields so mint genesis decodes under SDK v0.53.
func sanitizeMint(appState map[string]json.RawMessage) error {
	raw, ok := appState["mint"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var mint map[string]any
	if err := json.Unmarshal(raw, &mint); err != nil {
		return fmt.Errorf("unmarshal mint state: %w", err)
	}

	// minter: allow only inflation / annual_provisions
	if mraw, ok := mint["minter"]; ok && mraw != nil {
		if m, ok := mraw.(map[string]any); ok {
			allowed := map[string]struct{}{"inflation": {}, "annual_provisions": {}}
			for k := range m {
				if _, ok := allowed[k]; !ok {
					delete(m, k) // e.g. municipal_inflation
				}
			}
			mint["minter"] = m
		}
	}

	// params: keep a conservative whitelist
	if praw, ok := mint["params"]; ok && praw != nil {
		if p, ok := praw.(map[string]any); ok {
			allowed := map[string]struct{}{
				"mint_denom":            {},
				"inflation_rate_change": {},
				"inflation_min":         {},
				"inflation_max":         {},
				"goal_bonded":           {},
				"blocks_per_year":       {},
			}
			for k := range p {
				if k == "inflation_rate" || k == "municipal_inflation" {
					delete(p, k)
					continue
				}
				if _, ok := allowed[k]; !ok {
					delete(p, k)
				}
			}
			mint["params"] = p
		}
	}

	bz, err := json.Marshal(mint)
	if err != nil {
		return fmt.Errorf("marshal mint state: %w", err)
	}
	appState["mint"] = bz
	return nil
}

// ensureMintParamsSafe fills/repairs mint.params for SDK v0.53 to avoid div-by-zero in NextInflationRate.
func ensureMintParamsSafe(appState map[string]json.RawMessage, fallbackDenom string) error {
	raw, ok := appState["mint"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var st map[string]any
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("unmarshal mint: %w", err)
	}

	p, _ := st["params"].(map[string]any)
	if p == nil {
		p = map[string]any{}
	}

	isZeroDec := func(v any) bool {
		s, _ := v.(string)
		return s == "" || s == "0" || s == "0.0" || s == "0.000000000000000000"
	}

	// mint_denom
	if _, ok := p["mint_denom"]; !ok || p["mint_denom"] == "" {
		if rawSt, ok := appState["staking"]; ok {
			var stak map[string]any
			if json.Unmarshal(rawSt, &stak) == nil {
				if sp, ok := stak["params"].(map[string]any); ok {
					if den, _ := sp["bond_denom"].(string); den != "" {
						p["mint_denom"] = den
					}
				}
			}
		}
		if _, ok := p["mint_denom"]; !ok || p["mint_denom"] == "" {
			if fallbackDenom == "" {
				fallbackDenom = "afet"
			}
			p["mint_denom"] = fallbackDenom
		}
	}

	// goal_bonded > 0
	if _, ok := p["goal_bonded"]; !ok || isZeroDec(p["goal_bonded"]) {
		p["goal_bonded"] = "0.670000000000000000"
	}
	// inflation params
	if _, ok := p["inflation_rate_change"]; !ok || isZeroDec(p["inflation_rate_change"]) {
		p["inflation_rate_change"] = "0.130000000000000000"
	}
	if _, ok := p["inflation_min"]; !ok || isZeroDec(p["inflation_min"]) {
		p["inflation_min"] = "0.070000000000000000"
	}
	if _, ok := p["inflation_max"]; !ok || isZeroDec(p["inflation_max"]) {
		p["inflation_max"] = "0.200000000000000000"
	}
	// blocks_per_year
	switch bpv := p["blocks_per_year"].(type) {
	case string:
		if bpv == "" || bpv == "0" {
			p["blocks_per_year"] = "6311520" // ~5s blocks
		}
	case float64:
		if bpv <= 0 {
			p["blocks_per_year"] = "6311520"
		}
	default:
		p["blocks_per_year"] = "6311520"
	}

	st["params"] = p

	// Ensure minter exists with string fields
	m, _ := st["minter"].(map[string]any)
	if m == nil {
		m = map[string]any{}
	}
	if _, ok := m["inflation"]; !ok || m["inflation"] == "" {
		m["inflation"] = "0.100000000000000000"
	}
	if _, ok := m["annual_provisions"]; !ok || m["annual_provisions"] == "" {
		m["annual_provisions"] = "0.000000000000000000"
	}
	st["minter"] = m

	out, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshal mint: %w", err)
	}
	appState["mint"] = out
	return nil
}

// hasExpiredAllowance walks an allowance (possibly nested) and returns whether it has expired
// compared to genesisTime. If no expiration field is found, it returns (false, false).
func hasExpiredAllowance(v any, genesisTime time.Time) (expired bool, found bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return false, false
	}

	// direct expiration
	if expRaw, ok := m["expiration"]; ok {
		if s, ok := expRaw.(string); ok && s != "" {
			if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
				return !t.After(genesisTime), true
			}
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return !t.After(genesisTime), true
			}
			// malformed → treat as expired to be safe
			return true, true
		}
	}

	// nested allowances (different wrapper types)
	for _, key := range []string{"allowance", "basic", "grant"} {
		if v2, ok := m[key]; ok {
			if e, f := hasExpiredAllowance(v2, genesisTime); f {
				return e, true
			}
		}
	}

	return false, false
}

// sanitizeFeegrant removes any feegrant whose expiration ≤ genesisTime.
func sanitizeFeegrant(appState map[string]json.RawMessage, genesisTime time.Time) error {
	state, err := unmarshalState(appState, "feegrant")
	if err != nil {
		return fmt.Errorf("feegrant: %w", err)
	}

	allowances, _ := state["allowances"].([]any)
	if len(allowances) == 0 {
		return nil
	}

	filtered := make([]any, 0, len(allowances))
	for _, item := range allowances {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		allow, ok := m["allowance"]
		if !ok {
			continue
		}
		if expired, found := hasExpiredAllowance(allow, genesisTime); found && expired {
			continue // drop expired
		}
		filtered = append(filtered, m)
	}
	state["allowances"] = filtered

	if err := marshalState(appState, "feegrant", state); err != nil {
		return fmt.Errorf("feegrant marshal: %w", err)
	}
	return nil
}

// migrateIBCTransfer converts old transfer genesis:
// - denom_traces[{path, base_denom}]  -> denoms[{base, trace:[{port_id,channel_id}, ...]}]
// - adds params, port_id, total_escrowed (computed from escrow accounts)
// - removes denom_traces
func migrateIBCTransfer(appState map[string]json.RawMessage) error {
	const port = "transfer"

	transferState, err := unmarshalState(appState, "transfer")
	if err != nil {
		return fmt.Errorf("transfer: %w", err)
	}
	ibcState, _ := unmarshalState(appState, "ibc")
	bankState, _ := unmarshalState(appState, "bank")

	// 1) Build denoms from legacy denom_traces
	legacy, _ := transferState["denom_traces"].([]any)
	denoms := make([]any, 0, len(legacy))
	for _, it := range legacy {
		if m, ok := it.(map[string]any); ok {
			base, _ := m["base_denom"].(string)
			path, _ := m["path"].(string)
			denoms = append(denoms, map[string]any{
				"base":  base,
				"trace": pathToHops(path),
			})
		}
	}
	transferState["denoms"] = denoms
	delete(transferState, "denom_traces")

	// 2) Ensure required fields for v10
	transferState["port_id"] = port
	transferState["params"] = map[string]any{
		"send_enabled":    true,
		"receive_enabled": true,
	}

	// 3) Compute total_escrowed by summing balances of escrow accounts
	escrows := escrowAddresses(ibcState, port)
	totals := sumEscrow(bankState, escrows)
	transferState["total_escrowed"] = coinsJSON(totals)

	// 4) Write back
	return marshalState(appState, "transfer", transferState)
}

// "transfer/channel-0/transfer/channel-2" -> [{port_id:"transfer",channel_id:"channel-0"}, {port_id:"transfer",channel_id:"channel-2"}]
func pathToHops(path string) []any {
	if path == "" {
		return []any{}
	}
	parts := strings.Split(path, "/")
	hops := make([]any, 0, len(parts)/2)
	for i := 0; i+1 < len(parts); i += 2 {
		portID, chID := parts[i], parts[i+1]
		if portID == "" || chID == "" {
			continue
		}
		hops = append(hops, map[string]any{
			"port_id":    portID,
			"channel_id": chID,
		})
	}
	return hops
}

func escrowAddresses(ibcState map[string]any, port string) map[string]struct{} {
	out := make(map[string]struct{})
	chgen, _ := ibcState["channel_genesis"].(map[string]any)
	chans, _ := chgen["channels"].([]any)
	for _, it := range chans {
		ch, _ := it.(map[string]any)
		if ch == nil {
			continue
		}
		if ch["port_id"] != port {
			continue
		}
		chID, _ := ch["channel_id"].(string)
		if chID == "" {
			continue
		}
		addr := ibctransfertypes.GetEscrowAddress(port, chID) // AccAddress → String()
		out[addr.String()] = struct{}{}
	}
	return out
}

func sumEscrow(bankState map[string]any, escrows map[string]struct{}) map[string]*big.Int {
	out := map[string]*big.Int{}
	bals, _ := bankState["balances"].([]any)
	for _, it := range bals {
		b, _ := it.(map[string]any)
		if b == nil {
			continue
		}
		addr, _ := b["address"].(string)
		if _, ok := escrows[addr]; !ok {
			continue
		}
		coins, _ := b["coins"].([]any)
		for _, c := range coins {
			m, _ := c.(map[string]any)
			if m == nil {
				continue
			}
			denom, _ := m["denom"].(string)
			amt, _ := m["amount"].(string)
			if denom == "" || amt == "" {
				continue
			}
			if _, ok := out[denom]; !ok {
				out[denom] = new(big.Int)
			}
			if v, ok := new(big.Int).SetString(amt, 10); ok {
				out[denom].Add(out[denom], v)
			}
		}
	}
	return out
}

// migrateWasm updates wasm genesis for SDK/wasmd v0.53+/v0.61+:
//  1. Converts legacy AccessConfig {address:"..."} → {addresses:["..."]} in params and codes.*.
//  2. Removes legacy genesis message fields: gen_msgs / gen_mgs / genesis_msgs.
//  3. Renames contract_history → contract_code_history and ensures each contract
//     has at least one GENESIS entry using the provided initialHeight.
func migrateWasm(appState map[string]json.RawMessage, initialHeight uint64) error {
	// no-op if wasm not present
	if raw, ok := appState["wasm"]; !ok || len(raw) == 0 {
		return nil
	}

	state, err := unmarshalState(appState, "wasm")
	if err != nil {
		return fmt.Errorf("wasm: %w", err)
	}

	changed := false

	// 1) Fix AccessConfig shapes in params and codes
	if fixAccessConfigsInState(state) {
		changed = true
	}

	// 2) Drop legacy genesis message keys
	if deleteLegacyWasmGenesisMsgs(state) {
		changed = true
	}

	// 3) Normalize contract histories
	if normalizeContractHistories(state, initialHeight) {
		changed = true
	}

	if !changed {
		return nil
	}
	if err := marshalState(appState, "wasm", state); err != nil {
		return fmt.Errorf("wasm marshal: %w", err)
	}
	return nil
}

// --- helpers ---------------------------------------------------------------

// Convert {address:"..."} → {addresses:["..."]} (idempotent).
// Returns (newValue, changed).
func fixAccessConfigNode(v any) (any, bool) {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return v, false
	}

	changed := false

	// If "addresses" present, just remove legacy "address" if it exists.
	if _, hasList := m["addresses"]; hasList {
		if _, had := m["address"]; had {
			delete(m, "address")
			changed = true
		}
		return m, changed
	}

	// If legacy single "address" present, convert to "addresses".
	if addr, has := m["address"]; has {
		if s, ok := addr.(string); ok && s != "" {
			m["addresses"] = []any{s}
			changed = true
		}
		delete(m, "address")
		changed = true
	}

	return m, changed
}

// Fix AccessConfig wherever it appears in wasm genesis.
func fixAccessConfigsInState(state map[string]any) bool {
	changed := false

	// params: code_upload_access, instantiate_default_permission
	if pRaw, ok := state["params"]; ok && pRaw != nil {
		if p, ok := pRaw.(map[string]any); ok {
			if v, ok := p["code_upload_access"]; ok {
				if n, ch := fixAccessConfigNode(v); ch {
					p["code_upload_access"] = n
					changed = true
				}
			}
			if v, ok := p["instantiate_default_permission"]; ok {
				if n, ch := fixAccessConfigNode(v); ch {
					p["instantiate_default_permission"] = n
					changed = true
				}
			}
			state["params"] = p
		}
	}

	// codes[].code_info.instantiate_config
	if codesRaw, ok := state["codes"]; ok && codesRaw != nil {
		if arr, ok := codesRaw.([]any); ok {
			arrChanged := false
			for i, it := range arr {
				rec, ok := it.(map[string]any)
				if !ok {
					continue
				}
				if ciRaw, ok := rec["code_info"]; ok {
					if ci, ok := ciRaw.(map[string]any); ok {
						if ic, ok := ci["instantiate_config"]; ok {
							if n, ch := fixAccessConfigNode(ic); ch {
								ci["instantiate_config"] = n
								rec["code_info"] = ci
								arr[i] = rec
								arrChanged = true
							}
						}
					}
				}
			}
			if arrChanged {
				state["codes"] = arr
				changed = true
			}
		}
	}

	return changed
}

// Remove legacy wasm genesis message keys.
func deleteLegacyWasmGenesisMsgs(state map[string]any) bool {
	deleted := false
	for _, k := range []string{"gen_msgs", "gen_mgs", "genesis_msgs"} {
		if _, had := state[k]; had {
			delete(state, k)
			deleted = true
		}
	}
	return deleted
}

// Ensure each contract has non-empty contract_code_history, optionally rename legacy key.
func normalizeContractHistories(state map[string]any, initialHeight uint64) bool {
	contractsAny, ok := state["contracts"].([]any)
	if !ok || len(contractsAny) == 0 {
		return false
	}

	changed := false
	for i, it := range contractsAny {
		c, ok := it.(map[string]any)
		if !ok {
			continue
		}

		// Rename legacy key if present
		if oldHist, hasOld := c["contract_history"]; hasOld {
			if oldArr, ok := oldHist.([]any); ok {
				c["contract_code_history"] = oldArr
				changed = true
			}
			delete(c, "contract_history")
		}

		// Ensure non-empty contract_code_history
		histRaw, has := c["contract_code_history"]
		arr, ok := histRaw.([]any)
		if !has || !ok || len(arr) == 0 {
			c["contract_code_history"] = []any{makeGenesisHistoryEntry(c, initialHeight)}
			changed = true
		}

		contractsAny[i] = c
	}

	if changed {
		state["contracts"] = contractsAny
	}
	return changed
}

// Build a single GENESIS history record using contract code_id and initial height.
func makeGenesisHistoryEntry(contract map[string]any, initialHeight uint64) map[string]any {
	return map[string]any{
		"operation": "CONTRACT_CODE_HISTORY_OPERATION_TYPE_GENESIS",
		"code_id":   extractCodeID(contract),
		"updated": map[string]any{
			"block_height": fmt.Sprintf("%d", initialHeight),
			"tx_index":     "0",
		},
	}
}

// Try common places for code_id and stringify it.
func extractCodeID(contract map[string]any) string {
	// top-level code_id
	if v, ok := contract["code_id"]; ok {
		return stringifyJSONNumber(v)
	}
	// nested in contract_info
	if ciRaw, ok := contract["contract_info"]; ok {
		if ci, ok := ciRaw.(map[string]any); ok {
			if v, ok := ci["code_id"]; ok {
				return stringifyJSONNumber(v)
			}
		}
	}
	return "0"
}
