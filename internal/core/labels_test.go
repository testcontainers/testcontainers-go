package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeCustomLabels(t *testing.T) {
	t.Run("merge success", func(t *testing.T) {
		dst := map[string]string{"A": "1", "B": "2"}
		src := map[string]string{"B": "X", "C": "3"}

		err := MergeCustomLabels(dst, src)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"A": "1", "B": "X", "C": "3"}, dst)
	})

	t.Run("src cannot have keys starting with LabelBase", func(t *testing.T) {
		// --- Given ---
		dst := map[string]string{"A": "1", "B": "2"}
		src := map[string]string{"B": "X", LabelLang: "go"}

		// --- When ---
		err := MergeCustomLabels(dst, src)

		// --- Then ---
		require.EqualError(t, err, `cannot use prefix "org.testcontainers" for custom labels`)
		assert.Equal(t, map[string]string{"A": "1", "B": "2"}, dst)
	})
}
