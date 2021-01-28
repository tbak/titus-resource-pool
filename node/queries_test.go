package node

import (
	"github.com/Netflix/titus-resource-pool/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsKubletNode(t *testing.T) {
	kublet := util.ButNodeLabel(util.NewRandomNode(), NodeLabelBackend, NodeBackendKublet)
	require.True(t, IsKubletNode(kublet))

	notKublet := util.ButNodeLabel(util.NewRandomNode(), NodeLabelBackend, "TJC")
	require.False(t, IsKubletNode(notKublet))
}
