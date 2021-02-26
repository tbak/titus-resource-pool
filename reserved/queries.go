package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

func AsCapacityGroupReferenceList(capacityGroupList *capacityGroupV1.CapacityGroupList) []*capacityGroupV1.CapacityGroup {
	result := []*capacityGroupV1.CapacityGroup{}
	for _, node := range capacityGroupList.Items {
		tmp := node
		result = append(result, &tmp)
	}
	return result
}

func CapacityGroupResources(c *capacityGroupV1.CapacityGroup) poolV1.ComputeResource {
	return c.Spec.ComputeResource.Multiply(int64(c.Spec.InstanceCount))
}
