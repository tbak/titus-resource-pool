package resourcepool

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetResourcePoolMachineTypes(t *testing.T) {
	pool := ButResourcePoolMachineTypes(EmptyResourcePool(), []string{"r5.metal", "m5.metal"})
	require.Equal(t, []string{"r5.metal", "m5.metal"}, GetResourcePoolMachineTypes(pool))
}
