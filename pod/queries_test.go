package pod

import (
	"testing"

	"github.com/stretchr/testify/require"

	k8sCore "k8s.io/api/core/v1"

	"github.com/Netflix/titus-resource-pool/machine"
	"github.com/Netflix/titus-resource-pool/node"
	"github.com/Netflix/titus-resource-pool/util/xcollection"
)

var machineTypes = []string{"r5.metal", "m5.metal"}
var machineTypeSet = xcollection.SetOfStringList(machineTypes)

func TestGetPodRequestedMachineTypes(t *testing.T) {
	require.Equal(t, []string{}, GetPodRequestedMachineTypes(EmptyPod()))
	require.Equal(t,
		machineTypes,
		GetPodRequestedMachineTypes(ButPodMachineRequiredAffinity(EmptyPod(), machineTypes)),
	)
}

func TestFilterPodsOkWithMachineTypes(t *testing.T) {
	filtered := FilterPodsOkWithMachineTypes(
		[]*k8sCore.Pod{
			ButPodName(EmptyPod(), "pod1"),
			ButPodName(ButPodMachineRequiredAffinity(EmptyPod(), []string{"c5.metal"}), "pod2"),
			ButPodName(ButPodMachineRequiredAffinity(EmptyPod(), []string{"r5.metal", "c5.metal"}), "pod3"),
		},
		machineTypes,
	)
	require.True(t, len(filtered) == 2)
	require.Equal(t, "pod1", filtered[0].Name)
	require.Equal(t, "pod3", filtered[1].Name)
}

func TestIsPodOkWithMachineTypesSet(t *testing.T) {
	require.True(t, IsPodOkWithMachineTypesSet(EmptyPod(), machineTypeSet))
	require.True(t, IsPodOkWithMachineTypesSet(
		ButPodMachineRequiredAffinity(EmptyPod(), []string{"r5.metal", "c5.metal"}), machineTypeSet))
	require.False(t, IsPodOkWithMachineTypesSet(
		ButPodMachineRequiredAffinity(EmptyPod(), []string{"c5.metal"}), machineTypeSet))
}

func TestNotScheduledPodBelongsToResourcePool(t *testing.T) {
	pod := ButPodResourcePools(NewRandomNotScheduledPod(), "pool1, pool2")

	require.True(t, PodBelongsToResourcePool(pod, []string{"pool1", "pool2"}, "pool1", false, nil))
	require.True(t, PodBelongsToResourcePool(pod, []string{"pool1", "pool2"}, "pool2", false, nil))
	require.False(t, PodBelongsToResourcePool(pod, []string{"pool1", "pool2"}, "pool3", false, nil))
}

func TestScheduledPodBelongsToResourcePool(t *testing.T) {
	node1 := node.NewNode("node1", "pool1", machine.R5Metal())
	node2 := node.NewNode("node2", "pool2", machine.R5Metal())
	nodes := map[string]*k8sCore.Node{
		node1.Name: node1,
		node2.Name: node2,
	}

	pod := ButPodResourcePools(ButPodAssignedToNode(NewRandomNotScheduledPod(), node1), "pool1, pool2")

	require.True(t, PodBelongsToResourcePool(pod, []string{"pool1", "pool2"}, "pool1", false, nodes))
	require.False(t, PodBelongsToResourcePool(pod, []string{"pool1", "pool2"}, "pool2", false, nodes))
	require.False(t, PodBelongsToResourcePool(pod, []string{"pool1", "pool2"}, "pool3", false, nodes))
}
