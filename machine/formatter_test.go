package machine

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/Netflix/titus-resource-pool/util"
)

func TestFormatMachineTypeCompact(t *testing.T) {
	text := FormatMachineType(
		R5Metal(),
		FormatterOptions{Level: FormatCompact},
	)
	require.EqualValues(t,
		"{\"Name\":\"r5.metal\",\"ComputeResource\":{\"cpu\":96,\"gpu\":0,"+
			"\"memoryMB\":786432,\"diskMB\":1536000,\"networkMBPS\":25000}}",
		text,
	)
}
