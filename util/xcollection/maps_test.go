package xcollection

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetOfStringList(t *testing.T) {
	set := SetOfStringList([]string{"a", "b"})
	require.Contains(t, set, "a")
	require.Contains(t, set, "b")
	require.NotContains(t, set, "c")
}
