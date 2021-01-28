package resourcepool

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/node"
	"github.com/Netflix/titus-resource-pool/util"
	"github.com/stretchr/testify/require"
	k8sCore "k8s.io/api/core/v1"
	"testing"
)

func TestKubletNodesAreExcluded(t *testing.T) {
	pool := EmptyResourcePool()
	nodes := []*k8sCore.Node{
		util.NewNode("node1", pool.Name, util.R5Metal()),
		util.ButNodeLabel(util.NewNode("node2", pool.Name, util.R5Metal()), node.NodeLabelBackend, node.NodeBackendKublet),
	}

	// With kublet
	snapshot := NewStaticResourceSnapshot(pool, []*poolV1.MachineTypeConfig{}, nodes, []*k8sCore.Pod{},
		0, true)
	require.Equal(t, 2, len(snapshot.Nodes))
	require.Equal(t, 0, len(snapshot.ExcludedNodes))

	// Without kublet
	snapshot = NewStaticResourceSnapshot(pool, []*poolV1.MachineTypeConfig{}, nodes, []*k8sCore.Pod{},
		0, false)
	require.Equal(t, 1, len(snapshot.Nodes))
	require.Equal(t, 1, len(snapshot.ExcludedNodes))
}
