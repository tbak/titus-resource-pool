package node

import (
	"testing"
	"time"

	"github.com/Netflix/titus-resource-pool/machine"
	. "github.com/Netflix/titus-resource-pool/util"
	"github.com/stretchr/testify/require"
)

func TestFormatNodeCompact(t *testing.T) {
	text := FormatNode(
		NewNode("junitNode", "testResourcePool", machine.R5Metal()),
		10*time.Minute,
		FormatterOptions{Level: FormatCompact},
	)
	require.EqualValues(t,
		"{\"Name\":\"junitNode\",\"Up\":true,\"OnWayOut\":false}",
		text,
	)
}

func TestFormatNodeEssentials(t *testing.T) {
	text := FormatNode(
		NewNode("junitNode", "testResourcePool", machine.R5Metal()),
		10*time.Minute,
		FormatterOptions{Level: FormatEssentials},
	)
	require.EqualValues(t,
		"{\"Name\":\"junitNode\",\"Up\":true,\"OnWayOut\":false,\"AvailableResources\":{\"cpu\":96,\"gpu\":0,"+
			"\"memoryMB\":786432,\"diskMB\":1536000,\"networkMBPS\":25000}}",
		text,
	)
}
