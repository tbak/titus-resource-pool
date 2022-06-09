package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

func CapacityGroupResources(c *capacityGroupV1.CapacityGroup) poolV1.ComputeResource {
	return c.Spec.ComputeResource.Multiply(int64(c.Spec.InstanceCount))
}
