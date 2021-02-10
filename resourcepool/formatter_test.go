package resourcepool

import (
	"testing"

	"github.com/Netflix/titus-resource-pool/machine"
	. "github.com/Netflix/titus-resource-pool/util"
	"github.com/stretchr/testify/require"
)

func TestFormatResourcePoolCompact(t *testing.T) {
	text := FormatResourcePool(
		NewResourcePoolCrdOfMachine("unitTestPool", machine.R5Metal(), 4, 1),
		FormatterOptions{Level: FormatCompact},
	)
	require.EqualValues(t,
		"{\"Name\":\"unitTestPool\",\"ResourceCount\":1,\"AutoScalingEnabled\":true}",
		text,
	)
}

func TestFormatResourcePoolEssentials(t *testing.T) {
	text := FormatResourcePool(
		NewResourcePoolCrdOfMachine("unitTestPool", machine.R5Metal(), 4, 1),
		FormatterOptions{Level: FormatEssentials},
	)
	require.EqualValues(t,
		"{\"Name\":\"unitTestPool\",\"ResourceCount\":1,\"ResourceShape\":{\"cpu\":24,\"gpu\":0,"+
			"\"memoryMB\":196608,\"diskMB\":384000,\"networkMBPS\":6250},\"AutoScalingEnabled\":true}",
		text,
	)
}
