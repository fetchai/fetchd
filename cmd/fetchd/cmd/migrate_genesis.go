package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// migrate-genesis now supports initializing missing modules from ModuleBasics.
func MigrateGenesisCmd(basicManager module.BasicManager) *cobra.Command {
	var (
		stripModsCSV string
		initMissing  bool
		initOnlyCSV  string // if set, only initialize these modules (subset)
	)

	cmd := &cobra.Command{
		Use:   "migrate-genesis [old_genesis.json] [new_genesis.json]",
		Short: "Migrate a legacy (e.g., v0.47) genesis.json to be loadable by Cosmos SDK v0.53; can also init missing modules",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inFile := args[0]
			outFile := args[1]

			// Load old genesis
			appState, appGen, err := genutiltypes.GenesisStateFromGenFile(inFile)
			if err != nil {
				return fmt.Errorf("read genesis %q: %w", inFile, err)
			}

			// 1) Optionally strip legacy/removed modules
			for _, m := range parseCSV(stripModsCSV) {
				if _, ok := appState[m]; ok {
					delete(appState, m)
					fmt.Fprintf(cmd.ErrOrStderr(), "stripped module: %s\n", m)
				}
			}

			// 2) OPTIONAL: initialize missing modules with defaults from ModuleBasics
			if initMissing {
				// Build a JSON codec and get defaults for all modules declared by your app.
				ir := codectypes.NewInterfaceRegistry()
				basicManager.RegisterInterfaces(ir)
				cdc := codec.NewProtoCodec(ir)

				defaults := basicManager.DefaultGenesis(cdc)

				var initOnly = set(parseCSV(initOnlyCSV)) // optional whitelist
				// For each default, if missing in app_state, add it.
				for mod, def := range defaults {
					if _, exists := appState[mod]; exists {
						continue
					}
					// If a whitelist is provided, only init those listed.
					if len(initOnly) > 0 {
						if _, ok := initOnly[mod]; !ok {
							continue
						}
					}
					appState[mod] = def
					fmt.Fprintf(cmd.OutOrStdout(), "initialized missing module with default genesis: %s\n", mod)
				}
			}

			// 3) (Optional) targeted edits (e.g., Gov v1) — add your chain-specific conversion here.
			// if raw, ok := appState["gov"]; ok { /* adapt legacy fields if needed */ }
			if err := resetGovPreserveParamsAndFixBank(appState); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "stripped gov votes and deposits")

			if err := sanitizeMint(appState); err != nil {
				return err
			}

			gt, err := readGenesisTime(inFile) // inFile is your source genesis path
			if err != nil {
				return err
			}

			ih, err := readInitialHeight(inFile)
			if err != nil {
				return err
			}

			if err := sanitizeFeegrant(appState, gt); err != nil {
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

			if err := addGenesisHistoryToWasmContracts(appState, ih); err != nil {
				return err
			}

			// Write new app_state
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("marshal new app_state: %w", err)
			}
			appGen.AppState = appStateJSON

			// Persist
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
	cmd.Flags().StringVar(
		&stripModsCSV,
		"strip-modules",
		"capability,crisis",
		"Comma-separated module names to remove from app_state (legacy/unused)",
	)
	cmd.Flags().BoolVar(
		&initMissing,
		"init-missing-modules",
		true,
		"If true, initialize any missing modules using ModuleBasics.DefaultGenesis",
	)
	cmd.Flags().StringVar(
		&initOnlyCSV,
		"init-only",
		"",
		"Optional comma-separated allowlist of module names to initialize (implies --init-missing-modules)",
	)

	return cmd
}

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

// Reset Gov to a minimal v1 skeleton, but preserve basic params from the original genesis.
// Also zero the gov module account balance in bank and subtract it from total supply.
func resetGovPreserveParamsAndFixBank(appState map[string]json.RawMessage) error {
	// --- 1) Extract params from the original gov (v1 or legacy v1beta1) ---
	type coin struct{ Denom, Amount string }

	var (
		startingID = "1"
		paramsOut  = map[string]any{}
	)

	if rawGov, ok := appState["gov"]; ok && len(rawGov) > 0 {
		var old map[string]any
		if err := json.Unmarshal(rawGov, &old); err == nil {
			// keep starting_proposal_id if present
			if v, ok := old["starting_proposal_id"].(string); ok && v != "" {
				startingID = v
			}

			// v1-style params block?
			if p, ok := old["params"].(map[string]any); ok {
				for k, v := range p {
					paramsOut[k] = v // copy as-is
				}
			}

			// legacy split params?
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

	// If min_deposit missing, try to derive from staking bond denom or bank supply.
	ensureMinDeposit := func() {
		if _, ok := paramsOut["min_deposit"]; ok {
			return
		}
		// Try staking.params.bond_denom
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
		// Fallback: take first denom from bank.supply
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
		// Last resort: empty list (valid, but usually you’ll have a denom)
		paramsOut["min_deposit"] = []any{}
	}
	ensureMinDeposit()

	// Ensure some optional v1 params exist if not copied
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

	// Write minimal gov v1 state with preserved params
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

	// --- 2) Reconcile bank: zero gov module account and reduce supply accordingly ---
	govModAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	var bank map[string]any
	if rawBank, ok := appState["bank"]; ok && len(rawBank) > 0 {
		if err := json.Unmarshal(rawBank, &bank); err != nil {
			return fmt.Errorf("unmarshal bank: %w", err)
		}
	} else {
		return nil // nothing to fix
	}

	parseCoins := func(v any) ([]coin, bool) {
		arr, ok := v.([]any)
		if !ok {
			return nil, false
		}
		out := make([]coin, 0, len(arr))
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				den, _ := m["denom"].(string)
				amt, _ := m["amount"].(string)
				if den != "" && amt != "" {
					out = append(out, coin{Denom: den, Amount: amt})
				}
			}
		}
		return out, true
	}
	toAnyCoins := func(cs []coin) []any {
		out := make([]any, len(cs))
		for i, c := range cs {
			out[i] = map[string]any{"denom": c.Denom, "amount": c.Amount}
		}
		return out
	}

	// balances: zero gov module account coins
	var balances []any
	if v, ok := bank["balances"]; ok {
		if arr, ok := v.([]any); ok {
			balances = arr
		}
	}
	removed := []coin{}
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
		// Keep entry but zero coins (safer for indexers)
		m["coins"] = []any{}
		newBalances = append(newBalances, m)
	}
	bank["balances"] = newBalances

	// supply: subtract removed amounts (so supply stays consistent)
	var supply []coin
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
		newSupply := make([]coin, 0, len(cur))
		for den, v := range cur {
			newSupply = append(newSupply, coin{Denom: den, Amount: v.String()})
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

// sanitizeMint removes custom/unknown fields from mint.minter and mint.params
// so the state decodes under SDK v0.53. Specifically drops "municipal_inflation"
// from minter and "inflation_rate" (and other non-whitelisted keys) from params.
func sanitizeMint(appState map[string]json.RawMessage) error {
	raw, ok := appState["mint"]
	if !ok || len(raw) == 0 {
		return nil
	}

	var mint map[string]any
	if err := json.Unmarshal(raw, &mint); err != nil {
		return fmt.Errorf("unmarshal mint state: %w", err)
	}

	// --- sanitize minter (runtime values) ---
	if mraw, ok := mint["minter"]; ok && mraw != nil {
		if m, ok := mraw.(map[string]any); ok {
			// v0.53 Minter expects only these fields
			allowedMinter := map[string]struct{}{
				"inflation":         {},
				"annual_provisions": {},
			}
			for k := range m {
				if _, ok := allowedMinter[k]; !ok {
					// drop customs like "municipal_inflation"
					delete(m, k)
				}
			}
			mint["minter"] = m
		}
	}

	// --- sanitize params (configuration) ---
	if praw, ok := mint["params"]; ok && praw != nil {
		if p, ok := praw.(map[string]any); ok {
			// Keep a conservative whitelist. Adjust if your chain uses more fields.
			// Common SDK 0.5x params include these (depending on fork):
			allowedParams := map[string]struct{}{
				"mint_denom":            {},
				"inflation_rate_change": {}, // keep if present on your fork; remove if it still errors
				"inflation_min":         {},
				"inflation_max":         {},
				"goal_bonded":           {},
				"blocks_per_year":       {},
			}

			for k := range p {
				// explicitly drop known custom legacy keys
				if k == "inflation_rate" || k == "municipal_inflation" {
					delete(p, k)
					continue
				}
				// if you want to be strict, uncomment this block to drop anything non-whitelisted:
				if _, ok := allowedParams[k]; !ok {
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

func readGenesisTime(path string) (time.Time, error) {
	var hdr struct {
		GenesisTime string `json:"genesis_time"`
	}
	bz, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("read genesis: %w", err)
	}
	if err := json.Unmarshal(bz, &hdr); err != nil {
		return time.Time{}, fmt.Errorf("unmarshal header: %w", err)
	}
	if hdr.GenesisTime == "" {
		// fallback to Unix epoch if missing (shouldn’t be)
		return time.Unix(0, 0).UTC(), nil
	}
	t, err := time.Parse(time.RFC3339Nano, hdr.GenesisTime)
	if err != nil {
		// try looser RFC3339
		if t2, e2 := time.Parse(time.RFC3339, hdr.GenesisTime); e2 == nil {
			return t2.UTC(), nil
		}
		return time.Time{}, fmt.Errorf("parse genesis_time: %w", err)
	}
	return t.UTC(), nil
}

func readInitialHeight(path string) (uint64, error) {
	var hdr struct {
		InitialHeight string `json:"initial_height"`
	}
	bz, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read genesis: %w", err)
	}
	if err := json.Unmarshal(bz, &hdr); err != nil {
		return 0, fmt.Errorf("unmarshal header: %w", err)
	}
	if hdr.InitialHeight == "" {
		return 0, nil
	}
	u, err := strconv.ParseUint(hdr.InitialHeight, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse initial_height: %w", err)
	}
	return u, nil
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

	// Recursively check for expiration inside an allowance Any-JSON blob.
	var hasExpired func(v any) (bool, bool)
	hasExpired = func(v any) (expired bool, found bool) {
		m, ok := v.(map[string]any)
		if !ok {
			return false, false
		}
		// direct expiration?
		if expRaw, ok := m["expiration"]; ok {
			if s, ok := expRaw.(string); ok && s != "" {
				if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
					return !t.After(genesisTime), true
				}
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					return !t.After(genesisTime), true
				}
				// malformed -> treat as expired to be safe
				return true, true
			}
		}
		// nested: AllowedMsgAllowance{allowance}, PeriodicAllowance{basic}, etc.
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
		// Some forks nest under "grant": { "allowance": {...} }
		if v2, ok := m["grant"]; ok {
			if e, f := hasExpired(v2); f {
				return e, true
			}
		}
		return false, false
	}

	filtered := make([]any, 0, len(arrAny))
	dropped := 0
	for _, it := range arrAny {
		// structure: { "granter": "...", "grantee": "...", "allowance": { ...Any JSON... } }
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		allow, ok := m["allowance"]
		if !ok {
			// no allowance -> drop
			dropped++
			continue
		}
		if exp, found := hasExpired(allow); found && exp {
			dropped++
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
	if dropped > 0 {
		// optional: log via stdout/stderr in your command
		// fmt.Fprintf(os.Stderr, "feegrant: dropped %d expired grant(s)\n", dropped)
	}
	return nil
}

// sanitizeIBCTransfer removes legacy fields (e.g. "denom_traces") from IBC transfer genesis
// and ensures minimal required fields exist for ibc-go v10.
func sanitizeIBCTransfer(appState map[string]json.RawMessage) error {
	raw, ok := appState["transfer"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var gs map[string]any
	if err := json.Unmarshal(raw, &gs); err != nil {
		return fmt.Errorf("unmarshal transfer genesis: %w", err)
	}

	// legacy key to drop
	delete(gs, "denom_traces")

	// ibc-go v10 expects at least these:
	// - port_id (string, usually "transfer")
	// - params { send_enabled, receive_enabled }
	// - total_escrowed: []coin (can be empty)
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

// sanitizeWasmAccessConfigs fixes legacy AccessConfig objects that used "address"
// by converting them to "addresses":[address]. Safe to run multiple times.
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
		// if legacy singular "address" exists and "addresses" not set, convert
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

	// params: code_upload_access, instantiate_default_permission
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

// dropWasmGenesisMsgs removes legacy "gen_msgs" (and misspelt variants) from wasm genesis.
func dropWasmGenesisMsgs(appState map[string]json.RawMessage) error {
	raw, ok := appState["wasm"]
	if !ok || len(raw) == 0 {
		return nil
	}
	var st map[string]any
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("unmarshal wasm state: %w", err)
	}

	// Known legacy keys to remove
	delete(st, "gen_msgs")
	delete(st, "gen_mgs")      // some older chains used this typo
	delete(st, "genesis_msgs") // belt-and-suspenders

	// write back
	bz, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshal wasm state: %w", err)
	}
	appState["wasm"] = bz
	return nil
}

// addGenesisHistoryToWasmContracts renames `contract_history` -> `contract_code_history`
// and ensures each contract has at least one history entry. If missing, it appends a
// minimal GENESIS record using the provided initialHeight.
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

	// helper: normalize code_id to a JSON string (wasmd accepts stringified numbers)
	toCodeID := func(v any) any {
		switch x := v.(type) {
		case string:
			return x
		case float64:
			// JSON numbers come as float64; stringify without decimals
			return fmt.Sprintf("%.0f", x)
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

		// 1) If old key exists, move it to the new key and drop the old.
		if oldHistRaw, hasOld := c["contract_history"]; hasOld {
			if oldArr, ok := oldHistRaw.([]any); ok {
				c["contract_code_history"] = oldArr
			}
			delete(c, "contract_history")
			changed = true
		}

		// 2) Ensure contract_code_history exists and is non-empty.
		hRaw, hasNew := c["contract_code_history"]
		hArr, ok := hRaw.([]any)

		if !hasNew || !ok || len(hArr) == 0 {
			// find code_id (top-level or inside contract_info)
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
