package resourcepool

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/machine"
	"github.com/Netflix/titus-resource-pool/node"
	"github.com/stretchr/testify/require"
	k8sCore "k8s.io/api/core/v1"
	"testing"
)

func TestKubeletNodesAreExcluded(t *testing.T) {
	pool := EmptyResourcePool()
	nodes := []*k8sCore.Node{
		node.NewNode("node1", pool.Name, machine.R5Metal()),
		node.ButNodeLabel(node.NewNode("node2", pool.Name, machine.R5Metal()), node.NodeLabelBackend, node.NodeBackendKubelet),
	}

	// With kubelet
	snapshot := NewStaticResourceSnapshot(pool, []*poolV1.MachineTypeConfig{}, nodes, []*k8sCore.Pod{},
		0, true)
	require.Equal(t, 2, len(snapshot.Nodes))
	require.Equal(t, 0, len(snapshot.ExcludedNodes))

	// Without kubelet
	snapshot = NewStaticResourceSnapshot(pool, []*poolV1.MachineTypeConfig{}, nodes, []*k8sCore.Pod{},
		0, false)
	require.Equal(t, 1, len(snapshot.Nodes))
	require.Equal(t, 1, len(snapshot.ExcludedNodes))
}
