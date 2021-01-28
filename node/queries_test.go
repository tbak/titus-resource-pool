package node

import (
	"github.com/Netflix/titus-resource-pool/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsKubeletNode(t *testing.T) {
	kubelet := util.ButNodeLabel(util.NewRandomNode(), NodeLabelBackend, NodeBackendKubelet)
	require.True(t, IsKubeletNode(kubelet))

	notKubelet := util.ButNodeLabel(util.NewRandomNode(), NodeLabelBackend, "TJC")
	require.False(t, IsKubeletNode(notKubelet))
}
