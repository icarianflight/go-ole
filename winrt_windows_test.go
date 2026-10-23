//go:build windows

package ole

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHStringRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"ascii", "hello", "hello"},
		{"latin accents", "héllo wörld", "héllo wörld"},
		{"cjk", "日本語", "日本語"},
		// NewHString passes the rune count as the UTF-16 length, so every
		// surrogate pair costs one trailing code unit. Pinned, not endorsed.
		{"outside the BMP is truncated", "a😀b", "a😀"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := NewHString(tt.in)
			require.NoError(t, err)
			defer DeleteHString(h)

			assert.Equal(t, tt.want, h.String())
		})
	}
}
