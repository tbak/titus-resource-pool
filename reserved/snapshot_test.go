package reserved

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/resourcepool"

	"github.com/stretchr/testify/require"
	"testing"
)

func TestFindOwnByResourcePool(t *testing.T) {
	group1 := NewCapacityGroup("group1", resourcepool.PoolNameIntegration)
	group2 := NewCapacityGroup("group2", resourcepool.PoolNameIntegration2)

	snapshot := NewStaticCapacityGroupSnapshot([]*poolV1.CapacityGroup{group1, group2})
	require.Equal(t, 2, len(snapshot.CapacityGroups))

	integration := snapshot.FindOwnByResourcePool(resourcepool.PoolNameIntegration)
	require.Len(t, integration, 1)
	require.Equal(t, resourcepool.PoolNameIntegration, integration[0].Spec.ResourcePoolName)

	integration2 := snapshot.FindOwnByResourcePool(resourcepool.PoolNameIntegration2)
	require.Len(t, integration2, 1)
	require.Equal(t, resourcepool.PoolNameIntegration2, integration2[0].Spec.ResourcePoolName)
}
