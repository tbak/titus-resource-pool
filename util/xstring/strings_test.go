package xstring

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSplitByCommaAndTrim(t *testing.T) {
	require.Equal(t, []string{}, SplitByCommaAndTrim(""))
	require.Equal(t, []string{}, SplitByCommaAndTrim("   ,     "))
	require.Equal(t, []string{"a", "b"}, SplitByCommaAndTrim(" a , b ,,,"))
}
