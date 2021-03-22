package resourcepool

import (
	"testing"

	"github.com/stretchr/testify/require"

	k8sCore "k8s.io/api/core/v1"

	scaler "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/machine"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
)

func TestComputeAllocatableCapacity(t *testing.T) {
	node := poolNode.NewNode("node1", "myResourcePool", machine.R5Metal())
	nodeAvailable := machine.R5Metal().Spec.ComputeResource

	// Make CPU usage 75% while the rest is 50%
	podResources := nodeAvailable.Divide(2)
	podResources.CPU = podResources.CPU + podResources.CPU/2
	pod := poolPod.ButPodAssignedToNode(poolPod.ButPodResources(poolPod.NewRandomNotScheduledPod(), podResources), node)

	// We expect 25% across all dimensions
	nodes := map[string]*k8sCore.Node{node.Name: node}
	remaining := ComputeAllocatableCapacity(map[string]*k8sCore.Pod{pod.Name: pod}, nodes, scaler.Zero)
	require.Equal(t, nodeAvailable.Divide(4), remaining)
}
