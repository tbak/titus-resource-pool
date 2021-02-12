package xstring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitByCommaAndTrim(t *testing.T) {
	require.Equal(t, []string{}, SplitByCommaAndTrim(""))
	require.Equal(t, []string{}, SplitByCommaAndTrim("   ,     "))
	require.Equal(t, []string{"a", "b"}, SplitByCommaAndTrim(" a , b ,,,"))
}
