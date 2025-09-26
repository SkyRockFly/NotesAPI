package testutil

import (
	"bytes"
	"encoding/json"
	"testing"
)

func NormalizeJSON(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}
