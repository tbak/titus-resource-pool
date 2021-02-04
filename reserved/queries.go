package reserved

import (
	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

func AsCapacityGroupReferenceList(capacityGroupList *v1.CapacityGroupList) []*v1.CapacityGroup {
	result := []*v1.CapacityGroup{}
	for _, node := range capacityGroupList.Items {
		tmp := node
		result = append(result, &tmp)
	}
	return result
}

func CapacityGroupResources(c *v1.CapacityGroup) v1.ComputeResource {
	return c.Spec.ComputeResource.Multiply(int64(c.Spec.InstanceCount))
}
