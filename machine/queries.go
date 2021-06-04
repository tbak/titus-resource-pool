package machine

import machineV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"

func AsMachineTypeMap(machineTypes []*machineV1.MachineTypeConfig) map[string]*machineV1.MachineTypeConfig {
	result := map[string]*machineV1.MachineTypeConfig{}
	if machineTypes == nil {
		return result
	}
	for _, machineType := range machineTypes {
		result[machineType.Name] = machineType
	}
	return result
}
