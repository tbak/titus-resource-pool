package pod

import (
	"testing"

	"github.com/Netflix/titus-resource-pool/util/xcollection"
	"github.com/stretchr/testify/require"
	k8sCore "k8s.io/api/core/v1"
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
	require.True(t, IsPodOkWithMachineTypesSet(ButPodMachineRequiredAffinity(EmptyPod(), []string{"r5.metal", "c5.metal"}), machineTypeSet))
	require.False(t, IsPodOkWithMachineTypesSet(ButPodMachineRequiredAffinity(EmptyPod(), []string{"c5.metal"}), machineTypeSet))
}

