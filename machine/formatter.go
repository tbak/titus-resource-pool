package machine

import (
	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func FormatMachineType(machineType *machineTypeV1.MachineTypeConfig, options poolUtil.FormatterOptions) string {
	if options.Level != poolUtil.FormatDetails {
		return formatMachineTypeCompact(machineType)
	}
	return poolUtil.ToJSONString(machineType)
}

func formatMachineTypeCompact(machineType *machineTypeV1.MachineTypeConfig) string {
	type Compact struct {
		Name            string
		ComputeResource poolV1.ComputeResource
	}
	value := Compact{
		Name:            machineType.Name,
		ComputeResource: machineType.Spec.ComputeResource,
	}
	return poolUtil.ToJSONString(value)
}
