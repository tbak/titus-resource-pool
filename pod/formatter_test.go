package pod

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Netflix/titus-resource-pool/machine"
	"github.com/Netflix/titus-resource-pool/node"
	. "github.com/Netflix/titus-resource-pool/util"
)

func TestFormatPodCompact(t *testing.T) {
	text := FormatPod(
		ButPodName(NewRandomNotScheduledPod(), "testPod"),
		FormatterOptions{Level: FormatCompact},
	)
	require.EqualValues(t,
		"{\"Name\":\"testPod\",\"State\":\"notScheduled\",\"Node\":\"\"}",
		text,
	)
}

func TestFormatPodEssentials(t *testing.T) {
	text := FormatPod(
		ButPodRunningOnNode(ButPodName(NewRandomNotScheduledPod(), "testPod"),
			node.NewNode("junitNode", "testResourcePool", machine.R5Metal())),
		FormatterOptions{Level: FormatEssentials},
	)
	require.EqualValues(t,
		"{\"Name\":\"testPod\",\"State\":\"running\",\"Node\":\"junitNode\","+
			"\"ComputeResources\":{\"cpu\":24,\"gpu\":0,\"memoryMB\":196608,\"diskMB\":384000,\"networkMBPS\":6250}}",
		text,
	)
}
