package reserved

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	k8sCore "k8s.io/api/core/v1"

	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	"github.com/Netflix/titus-resource-pool/pod"
	"github.com/Netflix/titus-resource-pool/resourcepool"
	"github.com/Netflix/titus-resource-pool/util"
)

const (
	integrationBuffer = "buffer"
)

var (
	// Basic resource unit used in this test suite.
	computeUnit        = util.ComputeResourcesUnitProportional
	podShape           = computeUnit.Multiply(8)
	capacityGroupShape = computeUnit.Multiply(16)
	bufferShape        = computeUnit.Multiply(96)
)

func TestNewCapacityReservationUsageWithNoBuffer(t *testing.T) {
	metrics := NewUsageMetrics("test", resourcepool.PoolNameIntegration, integrationBuffer, true)

	poolSnapshot, pods := newResourcePoolSnapshotWithOneNodeAndScheduledPods(1)
	pod1 := pods[0]

	capacityGroup1, capacityGroup2, _ := newCapacityGroupsWithBuffer()
	capacityGroups := []*capacityGroupV1.CapacityGroup{capacityGroup1, capacityGroup2}

	usage := NewCapacityReservationUsage(poolSnapshot, capacityGroups, integrationBuffer)
	require.Len(t, usage.InCapacityGroup, 2)
	metrics.Update(usage)

	expectedGroup1Allocated := pod.FromPodToComputeResource(pod1).AlignResourceRatios(capacityGroup1.Spec.ComputeResource)
	expectedGroup1Unallocated := CapacityGroupResources(capacityGroups[0]).Sub(expectedGroup1Allocated)
	require.Equal(t, expectedGroup1Allocated, usage.InCapacityGroup["group-1"].Allocated)
	require.Equal(t, expectedGroup1Unallocated, usage.InCapacityGroup["group-1"].Unallocated)
	require.Equal(t, expectedGroup1Allocated, usage.AllReserved.Allocated)
	require.Equal(t, expectedGroup1Unallocated.Add(CapacityGroupResources(capacityGroups[1])),
		usage.AllReserved.Unallocated)
}

func TestNewCapacityReservationUsageWithNotUsedBuffer(t *testing.T) {
	metrics := NewUsageMetrics("test", resourcepool.PoolNameIntegration, integrationBuffer, true)

	poolSnapshot, pods := newResourcePoolSnapshotWithOneNodeAndScheduledPods(1)
	pod1 := pods[0]

	capacityGroup1, capacityGroup2, buffer := newCapacityGroupsWithBuffer()
	capacityGroups := []*capacityGroupV1.CapacityGroup{capacityGroup1, capacityGroup2, buffer}
	usage := NewCapacityReservationUsage(poolSnapshot, capacityGroups, integrationBuffer)
	require.Len(t, usage.InCapacityGroup, 2)
	metrics.Update(usage)

	bufferResources := CapacityGroupResources(buffer)

	expectedGroup1Allocated := pod.FromPodToComputeResource(pod1).AlignResourceRatios(capacityGroup1.Spec.ComputeResource)
	expectedGroup1Unallocated := CapacityGroupResources(capacityGroups[0]).Sub(expectedGroup1Allocated)
	require.Equal(t, expectedGroup1Allocated, usage.InCapacityGroup["group-1"].Allocated)
	require.Equal(t, expectedGroup1Unallocated, usage.InCapacityGroup["group-1"].Unallocated)
	require.Equal(t, expectedGroup1Allocated, usage.AllReserved.Allocated)
	require.Equal(t, expectedGroup1Unallocated.Add(CapacityGroupResources(capacityGroups[1])).Add(bufferResources),
		usage.AllReserved.Unallocated)
}

func TestNewCapacityReservationUsageWithUsedBuffer(t *testing.T) {
	metrics := NewUsageMetrics("test", resourcepool.PoolNameIntegration, integrationBuffer, true)

	// Capacity group size is 96 CPUs. Pod size is 8 CPUs. We need 12 pods to use reservation.
	poolSnapshot, pods := newResourcePoolSnapshotWithOneNodeAndScheduledPods(16)
	pod1 := pods[0]

	capacityGroup1, capacityGroup2, buffer := newCapacityGroupsWithBuffer()
	capacityGroups := []*capacityGroupV1.CapacityGroup{capacityGroup1, capacityGroup2, buffer}
	usage := NewCapacityReservationUsage(poolSnapshot, capacityGroups, integrationBuffer)
	require.Len(t, usage.InCapacityGroup, 2)
	metrics.Update(usage)

	podResources := pod.FromPodToComputeResource(pod1)
	expectedGroup1Allocated := podShape.Multiply(12)
	expectedGroup1Unallocated := poolV1.Zero

	bufferResources := CapacityGroupResources(buffer)
	expectedBufferAllocated := podResources.AlignResourceRatios(bufferResources).Multiply(4)
	expectedBufferUnallocated := CapacityGroupResources(buffer).Sub(expectedBufferAllocated)

	require.Equal(t, expectedGroup1Allocated, usage.InCapacityGroup["group-1"].Allocated)
	require.Equal(t, expectedGroup1Unallocated, usage.InCapacityGroup["group-1"].Unallocated)
	require.Equal(t, expectedGroup1Allocated.Add(expectedBufferAllocated), usage.AllReserved.Allocated)
	require.Equal(t, expectedGroup1Unallocated.Add(CapacityGroupResources(capacityGroups[1])).Add(expectedBufferUnallocated),
		usage.AllReserved.Unallocated)
}

func TestSameSubsystemDifferentResourcePool(t *testing.T) {
	const subsystem = "test_subsystem"
	const resourcePoolA = "resource_pool_a"
	const resourcePoolB = "resource_pool_b"

	_ = NewUsageMetrics(subsystem, resourcePoolA, integrationBuffer, true)
	_ = NewUsageMetrics(subsystem, resourcePoolB, integrationBuffer, true)
}

func newResourcePoolSnapshotWithOneNodeAndScheduledPods(podCount int) (*resourcepool.ResourceSnapshot, []*k8sCore.Pod) {
	pool := resourcepool.BasicResourcePool(resourcepool.PoolNameIntegration,
		20,
		computeUnit.Multiply(96),
	)

	node := poolNode.NewNode("node1", resourcepool.PoolNameIntegration, util.MachineFromUnitProportional96())

	// '_' chars in the capacity-group pod label/annotation value will be converted to '-' to
	// find the pod's corresponding Capacity Group CRD in kube
	pods := []*k8sCore.Pod{}
	for i := 0; i < podCount; i++ {
		newPod := pod.NewNotScheduledPod(resourcepool.PoolNameIntegration, podShape, time.Now())
		newPod = pod.ButPodCapacityGroup(newPod, "group_1")
		newPod = pod.ButPodAssignedToNode(newPod, node)
		pods = append(pods, newPod)
	}

	return resourcepool.NewStaticResourceSnapshot(pool, []*machineTypeV1.MachineTypeConfig{}, []*k8sCore.Node{node},
		pods, 0, 0, true), pods
}

func newCapacityGroupsWithBuffer() (*capacityGroupV1.CapacityGroup, *capacityGroupV1.CapacityGroup, *capacityGroupV1.CapacityGroup) {
	capacityGroup1 := BasicCapacityGroup("group-1", resourcepool.PoolNameIntegration, capacityGroupShape, 6)
	capacityGroup2 := BasicCapacityGroup("group2", resourcepool.PoolNameIntegration, capacityGroupShape, 6)
	buffer := BasicCapacityGroup(integrationBuffer, resourcepool.PoolNameIntegration, bufferShape, 1)
	return capacityGroup1, capacityGroup2, buffer
}
