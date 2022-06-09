package util

import (
	v1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
)

func GetNormalizedCapacityGroupName(c *v1.CapacityGroup) string {
	if original, ok := c.ObjectMeta.Annotations["capacitygroup.com.netflix.titus/original-name"]; ok {
		return original
	}
	return c.ObjectMeta.Name
}
