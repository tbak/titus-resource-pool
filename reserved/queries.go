package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

func CapacityGroupResources(c *capacityGroupV1.CapacityGroup) poolV1.ComputeResource {
	return c.Spec.ComputeResource.Multiply(int64(c.Spec.InstanceCount))
}

func GetNormalizedCapacityGroupName(c *capacityGroupV1.CapacityGroup) string {
	if original, ok := c.ObjectMeta.Annotations["capacitygroup.com.netflix.titus/original-name"]; ok {
		return original
	}
	return c.ObjectMeta.Name
}
