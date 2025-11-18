package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

// AddGenesisWasmMsgCmd provides a compatible command group even after wasmd removed the old genesis-msg CLI.
// It offers:
//   - list-codes: print wasm.codes from genesis
//   - list-contracts: print wasm.contracts from genesis
//   - merge: merge a JSON fragment into app_state.wasm (declarative genesis)
//
// The old commands store-code / instantiate / execute are kept as stubs that explain what to do now.
func AddGenesisWasmMsgCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "add-wasm-genesis-message",
		Short:                      "Wasm genesis helpers (compat layer for removed wasmd genesis cmds)",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		newListCodesCmd(),
		newListContractsCmd(),
		newMergeWasmGenesisCmd(),
		newStubCmd("store-code", "Removed upstream. Use `merge` to declare codes in genesis or use post-genesis tx: `wasmd tx wasm store ...`"),
		newStubCmd("instantiate-contract", "Removed upstream. Declare contracts in genesis via `merge`, or instantiate post-genesis via tx."),
		newStubCmd("execute-contract", "Removed upstream. Execution in genesis is no longer supported; use a tx after chain start."),
	)

	return cmd
}

// ---- list-codes ----

func newListCodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-codes",
		Short: "List wasm codes from genesis (app_state.wasm.codes)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config
			cfg.SetRoot(clientCtx.HomeDir)

			appState, appGen, err := genutiltypes.GenesisStateFromGenFile(cfg.GenesisFile())
			if err != nil {
				return fmt.Errorf("read genesis: %w", err)
			}

			wasmRaw := appState["wasm"]
			if len(wasmRaw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no wasm module in app_state")
				return nil
			}

			var wasm map[string]json.RawMessage
			if err := json.Unmarshal(wasmRaw, &wasm); err != nil {
				return fmt.Errorf("unmarshal wasm app_state: %w", err)
			}

			var codes []struct {
				CodeID    uint64          `json:"code_id"`
				CodeInfo  json.RawMessage `json:"code_info"`
				CodeBytes string          `json:"code_bytes,omitempty"`
			}
			if raw := wasm["codes"]; len(raw) > 0 {
				if err := json.Unmarshal(raw, &codes); err != nil {
					return fmt.Errorf("unmarshal wasm.codes: %w", err)
				}
			}

			if len(codes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no codes found")
				return nil
			}

			sort.Slice(codes, func(i, j int) bool { return codes[i].CodeID < codes[j].CodeID })

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "GENESIS_HEIGHT:\t%d\n", appGen.InitialHeight)
			fmt.Fprintln(w, "CODE_ID")
			for _, c := range codes {
				fmt.Fprintf(w, "%d\n", c.CodeID)
			}
			return w.Flush()
		},
	}
}

// ---- list-contracts ----

func newListContractsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-contracts",
		Short: "List wasm contracts from genesis (app_state.wasm.contracts)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config
			cfg.SetRoot(clientCtx.HomeDir)

			appState, appGen, err := genutiltypes.GenesisStateFromGenFile(cfg.GenesisFile())
			if err != nil {
				return fmt.Errorf("read genesis: %w", err)
			}

			wasmRaw := appState["wasm"]
			if len(wasmRaw) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no wasm module in app_state")
				return nil
			}

			var wasm map[string]json.RawMessage
			if err := json.Unmarshal(wasmRaw, &wasm); err != nil {
				return fmt.Errorf("unmarshal wasm app_state: %w", err)
			}

			var contracts []struct {
				ContractAddress string            `json:"contract_address"`
				ContractInfo    json.RawMessage   `json:"contract_info"`
				ContractState   []json.RawMessage `json:"contract_state,omitempty"`
			}
			if raw := wasm["contracts"]; len(raw) > 0 {
				if err := json.Unmarshal(raw, &contracts); err != nil {
					return fmt.Errorf("unmarshal wasm.contracts: %w", err)
				}
			}

			if len(contracts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no contracts found")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "GENESIS_HEIGHT:\t%d\n", appGen.InitialHeight)
			fmt.Fprintln(w, "CONTRACT_ADDRESS")
			for _, c := range contracts {
				fmt.Fprintf(w, "%s\n", c.ContractAddress)
			}
			return w.Flush()
		},
	}
}

// ---- merge (declarative) ----

// newMergeWasmGenesisCmd merges a JSON fragment into app_state.wasm.
// The input file should contain keys like: "codes", "contracts", "sequences", "pinned_codes", "params".
// Example:
//
//	{
//	  "codes": [{ "code_id": 1, "code_info": {...}, "code_bytes": "<base64>" }],
//	  "contracts": [{ "contract_address": "...", "contract_info": {...}, "contract_state": [...] }]
//	}
func newMergeWasmGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge [wasm_genesis_fragment.json]",
		Short: "Merge a wasm genesis JSON fragment into app_state.wasm (supported replacement for old genesis-msg cmds)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config
			cfg.SetRoot(clientCtx.HomeDir)

			genFile := cfg.GenesisFile()

			appState, appGen, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("read genesis: %w", err)
			}

			// read fragment
			in := args[0]
			bz, err := os.ReadFile(in)
			if err != nil {
				return fmt.Errorf("read fragment: %w", err)
			}

			var fragment map[string]json.RawMessage
			if err := json.Unmarshal(bz, &fragment); err != nil {
				return fmt.Errorf("unmarshal fragment: %w", err)
			}
			if len(fragment) == 0 {
				return errors.New("fragment is empty")
			}

			// current wasm state (object)
			var wasm map[string]json.RawMessage
			if raw := appState["wasm"]; len(raw) > 0 {
				if err := json.Unmarshal(raw, &wasm); err != nil {
					return fmt.Errorf("unmarshal existing wasm state: %w", err)
				}
			}
			if wasm == nil {
				wasm = make(map[string]json.RawMessage)
			}

			// merge keys (shallow merge: override same keys, add new keys)
			for k, v := range fragment {
				wasm[k] = v
			}

			// write back wasm object
			wasmBz, err := json.Marshal(wasm)
			if err != nil {
				return fmt.Errorf("marshal merged wasm state: %w", err)
			}
			appState["wasm"] = wasmBz

			// re-encode full app_state into the app genesis
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("marshal application state: %w", err)
			}
			appGen.AppState = appStateJSON

			if err := genutil.ExportGenesisFile(appGen, genFile); err != nil {
				return fmt.Errorf("write genesis: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "merged %d key(s) into app_state.wasm: %s\n", len(fragment), strings.Join(mapKeys(fragment), ", "))
			return nil
		},
	}
	return cmd
}

// ---- deprecation stubs ----

func newStubCmd(use, msg string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: "(stub) This genesis-msg was removed upstream",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.ErrOrStderr(), "Command %q is no longer supported at genesis.\n%s\n", use, msg)
			return fmt.Errorf("unsupported genesis command: %s", use)
		},
	}
}

// ---- helpers ----

func mapKeys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
