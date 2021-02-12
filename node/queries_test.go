package node

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	k8sCore "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsKubeletNode(t *testing.T) {
	kubelet := ButNodeLabel(NewRandomNode(), NodeLabelBackend, NodeBackendKubelet)
	require.True(t, IsKubeletNode(kubelet))

	notKubelet := ButNodeLabel(NewRandomNode(), NodeLabelBackend, "TJC")
	require.False(t, IsKubeletNode(notKubelet))
}

func TestSortNodesByAge(t *testing.T) {
	newNodeCreatedAt := func(timestamp time.Time) *k8sCore.Node {
		return NewRandomNode(func(node *k8sCore.Node) { node.CreationTimestamp = metaV1.Time{Time: timestamp} })
	}

	now := time.Now()
	nodes := []*k8sCore.Node{
		newNodeCreatedAt(now.Add(-0 * time.Hour)),
		newNodeCreatedAt(now.Add(-1 * time.Hour)),
		newNodeCreatedAt(now.Add(-3 * time.Hour)),
		newNodeCreatedAt(now.Add(-2 * time.Hour)),
	}
	sorted := SortNodesByAge(nodes)
	require.EqualValues(t, sorted, nodes) // Sorting in place
	require.EqualValues(t, now.Add(-3*time.Hour), sorted[0].CreationTimestamp.Time)
	require.EqualValues(t, now.Add(-2*time.Hour), sorted[1].CreationTimestamp.Time)
	require.EqualValues(t, now.Add(-1*time.Hour), sorted[2].CreationTimestamp.Time)
	require.EqualValues(t, now.Add(-0*time.Hour), sorted[3].CreationTimestamp.Time)
}
