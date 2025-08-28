package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
)

func unmarshalState(app map[string]json.RawMessage, key string) (map[string]any, error) {
	raw := app[key]
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var st map[string]any
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, err
	}
	return st, nil
}

func marshalState(app map[string]json.RawMessage, key string, st map[string]any) error {
	bz, err := json.Marshal(st)
	if err != nil {
		return err
	}
	app[key] = bz
	return nil
}

func coinsJSON(totals map[string]*big.Int) []any {
	out := make([]any, 0, len(totals))
	for den, v := range totals {
		out = append(out, map[string]any{"denom": den, "amount": v.String()})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].(map[string]any)["denom"].(string) <
			out[j].(map[string]any)["denom"].(string)
	})
	return out
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

func toStringSet(list []string) map[string]struct{} {
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

func stringifyJSONNumber(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%.0f", x)
	case json.Number:
		return x.String()
	default:
		return "0"
	}
}
