package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuoteIdentifier(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  string
	}{
		{"database", `"database"`},
		{`a"b"`, `"a""b"""`},
		{`a'b\`, `"a'b\"`},
		{"日本語", `"日本語"`},
	} {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.want, quoteIdentifier(tt.input))
		})
	}
}

func TestQuoteLiteral(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  string
	}{
		{"database", `'database'`},
		{"", `''`},
		{`a'b'`, `'a''b'''`},
		{`a"b`, `'a"b'`},
		{`a'b\`, `E'a''b\\'`},
	} {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.want, quoteLiteral(tt.input))
		})
	}
}
