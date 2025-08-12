package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
				only := set(parseCSV(initOnlyCSV))

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
			if err := sanitizeIBCTransfer(appState); err != nil {
				return err
			}
			if err := sanitizeWasmAccessConfigs(appState); err != nil {
				return err
			}
			if err := dropWasmGenesisMsgs(appState); err != nil {
				return err
			}
			if err := addGenesisHistoryToWasmContracts(appState, initHeight); err != nil {
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

// ---------------------------- helpers ----------------------------

func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func set(list []string) map[string]struct{} {
	if len(list) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(list))
	for _, v := range list {
		m[v] = struct{}{}
	}
	return m
}

func ensureDir(path string) error {
	dir := ""
	if i := strings.LastIndex(path, "/"); i >= 0 {
		dir = path[:i]
	}
	if dir == "" {
		return nil
	}
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create dir %q: %w", dir, err)
	}
	return nil
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

// sanitizeFeegrant removes any grant whose (nested) expiration ≤ genesisTime.
func sanitizeFeegrant(appState map[string]json.RawMessage, genesisTime time.Time) error {
	raw, ok := appState["feegrant"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var st map[string]any
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("unmarshal feegrant: %w", err)
	}

	arrAny, _ := st["allowances"].([]any)
	if len(arrAny) == 0 {
		return nil
	}

	var hasExpired func(v any) (expired bool, found bool)
	hasExpired = func(v any) (bool, bool) {
		m, ok := v.(map[string]any)
		if !ok {
			return false, false
		}
		if expRaw, ok := m["expiration"]; ok {
			if s, ok := expRaw.(string); ok && s != "" {
				if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
					return !t.After(genesisTime), true
				}
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					return !t.After(genesisTime), true
				}
				return true, true // malformed → drop
			}
		}
		if v2, ok := m["allowance"]; ok {
			if e, f := hasExpired(v2); f {
				return e, true
			}
		}
		if v2, ok := m["basic"]; ok {
			if e, f := hasExpired(v2); f {
				return e, true
			}
		}
		if v2, ok := m["grant"]; ok {
			if e, f := hasExpired(v2); f {
				return e, true
			}
		}
		return false, false
	}

	filtered := make([]any, 0, len(arrAny))
	for _, it := range arrAny {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		allow, ok := m["allowance"]
		if !ok {
			continue
		}
		if exp, found := hasExpired(allow); found && exp {
			continue
		}
		filtered = append(filtered, m)
	}
	st["allowances"] = filtered

	nbz, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshal feegrant: %w", err)
	}
	appState["feegrant"] = nbz
	return nil
}

// sanitizeIBCTransfer removes legacy fields and ensures minimal required fields for ibc-go v10.
func sanitizeIBCTransfer(appState map[string]json.RawMessage) error {
	raw, ok := appState["transfer"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var gs map[string]any
	if err := json.Unmarshal(raw, &gs); err != nil {
		return fmt.Errorf("unmarshal transfer genesis: %w", err)
	}

	delete(gs, "denom_traces")

	if _, ok := gs["port_id"]; !ok {
		gs["port_id"] = "transfer"
	}
	if p, ok := gs["params"].(map[string]any); ok {
		if _, ok := p["send_enabled"]; !ok {
			p["send_enabled"] = true
		}
		if _, ok := p["receive_enabled"]; !ok {
			p["receive_enabled"] = true
		}
		gs["params"] = p
	} else {
		gs["params"] = map[string]any{
			"send_enabled":    true,
			"receive_enabled": true,
		}
	}
	if _, ok := gs["total_escrowed"]; !ok {
		gs["total_escrowed"] = []any{}
	}

	bz, err := json.Marshal(gs)
	if err != nil {
		return fmt.Errorf("marshal transfer genesis: %w", err)
	}
	appState["transfer"] = bz
	return nil
}

// sanitizeWasmAccessConfigs converts legacy AccessConfig {address:"..."} to {addresses:["..."]}.
func sanitizeWasmAccessConfigs(appState map[string]json.RawMessage) error {
	raw, ok := appState["wasm"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var wasm map[string]any
	if err := json.Unmarshal(raw, &wasm); err != nil {
		return fmt.Errorf("unmarshal wasm state: %w", err)
	}

	fixAC := func(v any) any {
		m, ok := v.(map[string]any)
		if !ok || m == nil {
			return v
		}
		if _, hasList := m["addresses"]; !hasList {
			if addr, has := m["address"]; has {
				if s, ok := addr.(string); ok && s != "" {
					m["addresses"] = []any{s}
				}
				delete(m, "address")
			}
		}
		return m
	}

	// params
	if pRaw, ok := wasm["params"]; ok && pRaw != nil {
		if p, ok := pRaw.(map[string]any); ok {
			if cua, ok := p["code_upload_access"]; ok {
				p["code_upload_access"] = fixAC(cua)
			}
			if idp, ok := p["instantiate_default_permission"]; ok {
				p["instantiate_default_permission"] = fixAC(idp)
			}
			wasm["params"] = p
		}
	}
	// codes[].code_info.instantiate_config
	if codesRaw, ok := wasm["codes"]; ok && codesRaw != nil {
		if arr, ok := codesRaw.([]any); ok {
			for i, it := range arr {
				rec, ok := it.(map[string]any)
				if !ok {
					continue
				}
				if ciRaw, ok := rec["code_info"]; ok {
					if ci, ok := ciRaw.(map[string]any); ok {
						if ic, ok := ci["instantiate_config"]; ok {
							ci["instantiate_config"] = fixAC(ic)
							rec["code_info"] = ci
							arr[i] = rec
						}
					}
				}
			}
			wasm["codes"] = arr
		}
	}

	bz, err := json.Marshal(wasm)
	if err != nil {
		return fmt.Errorf("marshal wasm state: %w", err)
	}
	appState["wasm"] = bz
	return nil
}

// dropWasmGenesisMsgs removes legacy wasm "gen_msgs" variants.
func dropWasmGenesisMsgs(appState map[string]json.RawMessage) error {
	raw, ok := appState["wasm"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var st map[string]any
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("unmarshal wasm state: %w", err)
	}

	delete(st, "gen_msgs")
	delete(st, "gen_mgs")
	delete(st, "genesis_msgs")

	bz, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshal wasm state: %w", err)
	}
	appState["wasm"] = bz
	return nil
}

// addGenesisHistoryToWasmContracts renames contract_history -> contract_code_history
// and ensures each contract has at least one GENESIS history entry.
func addGenesisHistoryToWasmContracts(appState map[string]json.RawMessage, initialHeight uint64) error {
	raw, ok := appState["wasm"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var st map[string]any
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("unmarshal wasm state: %w", err)
	}

	contractsAny, ok := st["contracts"].([]any)
	if !ok || len(contractsAny) == 0 {
		return nil
	}

	toCodeID := func(v any) any {
		switch x := v.(type) {
		case string:
			return x
		case float64:
			return fmt.Sprintf("%.0f", x) // stringify JSON number
		case json.Number:
			return x.String()
		default:
			return "0"
		}
	}

	genesisHist := func(codeID any) map[string]any {
		return map[string]any{
			"operation": "CONTRACT_CODE_HISTORY_OPERATION_TYPE_GENESIS",
			"code_id":   toCodeID(codeID),
			"updated": map[string]any{
				"block_height": fmt.Sprintf("%d", initialHeight),
				"tx_index":     "0",
			},
		}
	}

	changed := false
	for i, it := range contractsAny {
		c, ok := it.(map[string]any)
		if !ok {
			continue
		}
		// rename if old key
		if oldHistRaw, hasOld := c["contract_history"]; hasOld {
			if oldArr, ok := oldHistRaw.([]any); ok {
				c["contract_code_history"] = oldArr
			}
			delete(c, "contract_history")
			changed = true
		}
		// ensure non-empty
		hRaw, hasNew := c["contract_code_history"]
		hArr, ok := hRaw.([]any)
		if !hasNew || !ok || len(hArr) == 0 {
			var codeID any
			if v, ok := c["code_id"]; ok {
				codeID = v
			} else if ci, ok := c["contract_info"].(map[string]any); ok {
				codeID = ci["code_id"]
			}
			if codeID == nil {
				codeID = "0"
			}
			c["contract_code_history"] = []any{genesisHist(codeID)}
			changed = true
		}
		contractsAny[i] = c
	}

	if changed {
		st["contracts"] = contractsAny
		bz, err := json.Marshal(st)
		if err != nil {
			return fmt.Errorf("marshal wasm state: %w", err)
		}
		appState["wasm"] = bz
	}
	return nil
}
