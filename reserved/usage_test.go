package reserved

import (
	"github.com/stretchr/testify/require"
	"testing"

	k8sCore "k8s.io/api/core/v1"

	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	"github.com/Netflix/titus-resource-pool/machine"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	"github.com/Netflix/titus-resource-pool/pod"
	"github.com/Netflix/titus-resource-pool/resourcepool"
)

func TestNewCapacityReservationUsage(t *testing.T) {
	pool := resourcepool.ButResourcePoolName(resourcepool.EmptyResourcePool(), resourcepool.PoolNameIntegration)
	pool.Spec.ResourceCount = 20

	node := poolNode.NewNode("node1", resourcepool.PoolNameIntegration, machine.R5Metal())

	pod1 := pod.ButPodResourcePools(pod.NewRandomNotScheduledPod(), resourcepool.PoolNameIntegration)
	// '_' chars in the capacity-group pod label/annotation value will be converted to '-' to
	// find the pod's correponding Capacity Group CRD in kube
	pod1 = pod.ButPodCapacityGroup(pod1, "group_1")
	pod1 = pod.ButPodAssignedToNode(pod1, node)

	poolSnapshot := resourcepool.NewStaticResourceSnapshot(pool, []*machineTypeV1.MachineTypeConfig{}, []*k8sCore.Node{node},
		[]*k8sCore.Pod{pod1}, 0, true)

	group1 := NewCapacityGroup("group-1", resourcepool.PoolNameIntegration)
	group1.Spec.InstanceCount = 10
	capacityGroups := []*capacityGroupV1.CapacityGroup{
		group1,
		NewCapacityGroup("group2", resourcepool.PoolNameIntegration),
	}

	usage := NewCapacityReservationUsage(poolSnapshot, capacityGroups)
	require.Len(t, usage.InCapacityGroup, 2)

	expectedGroup1Allocated := pod.FromPodToComputeResource(pod1)
	expectedGroup1Unallocated := CapacityGroupResources(capacityGroups[0]).Sub(expectedGroup1Allocated)
	require.Equal(t, expectedGroup1Allocated, usage.InCapacityGroup["group-1"].Allocated)
	require.Equal(t, expectedGroup1Unallocated, usage.InCapacityGroup["group-1"].Unallocated)
	require.Equal(t, expectedGroup1Allocated, usage.AllReserved.Allocated)
	require.Equal(t, expectedGroup1Unallocated.Add(CapacityGroupResources(capacityGroups[1])),
		usage.AllReserved.Unallocated)
}
