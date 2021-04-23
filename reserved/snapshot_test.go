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

func TestFilterCapacityGroups(t *testing.T) {
	noAnnotationsGroup := NewCapacityGroup("noAnnotation", "noAnnotationPool")

	criticalKube := NewCapacityGroup("criticalKube", "criticalKubePool")
	criticalKube.Annotations = map[string]string{
		"tier": "Critical",
	}
	criticalKube.Spec.SchedulerName = PodSchedulerKube

	criticalFenzo := NewCapacityGroup("criticalFenzo", "criticalFenzoPool")
	criticalKube.Annotations = map[string]string{
		"tier": "Critical",
	}
	criticalFenzo.Spec.SchedulerName = PodSchedulerFenzo

	flexKube:= NewCapacityGroup("flexKube", "flexKubePool")
	flexKube.Annotations = map[string]string{
		"tier": "Flex",
	}
	flexKube.Spec.SchedulerName = PodSchedulerKube

	flexFenzo:= NewCapacityGroup("flexFenzo", "flexFenzoPool")
	flexFenzo.Annotations = map[string]string{
		"tier": "Flex",
	}
	flexKube.Spec.SchedulerName = PodSchedulerFenzo

	cgList := capacityGroupV1.CapacityGroupList{
		Items: []capacityGroupV1.CapacityGroup{
			*noAnnotationsGroup,
			*criticalKube,
			*criticalFenzo,
			*flexKube,
			*flexFenzo,
		},
	}

	// Expect only critical tier with kubescheduler CGs
	filteredCGs := filterCapacityGroups(cgList)
	require.Len(t, filteredCGs, 2)
	require.Equal(t, noAnnotationsGroup.Name, filteredCGs[0].Name)
	require.Equal(t, criticalKube.Name, filteredCGs[1].Name)
}
