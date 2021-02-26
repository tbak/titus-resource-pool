package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV2 "github.com/Netflix/titus-resource-pool/resourcepool"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindOwnByResourcePool(t *testing.T) {
	group1 := NewCapacityGroup("group1", poolV2.PoolNameIntegration)
	group2 := NewCapacityGroup("group2", poolV2.PoolNameIntegration2)

	snapshot := NewStaticCapacityGroupSnapshot([]*capacityGroupV1.CapacityGroup{group1, group2})
	require.Equal(t, 2, len(snapshot.CapacityGroups))

	integration := snapshot.FindOwnedByResourcePool(poolV2.PoolNameIntegration)
	require.Len(t, integration, 1)
	require.Equal(t, poolV2.PoolNameIntegration, integration[0].Spec.ResourcePoolName)

	integration2 := snapshot.FindOwnedByResourcePool(poolV2.PoolNameIntegration2)
	require.Len(t, integration2, 1)
	require.Equal(t, poolV2.PoolNameIntegration2, integration2[0].Spec.ResourcePoolName)
}
