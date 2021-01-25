package resourcepool

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/util/xstring"
)

// Machine types used by a resource pool or empty array if none is defined.
func GetResourcePoolMachineTypes(resourcePool *poolV1.ResourcePoolConfig) []string {
	value, ok := resourcePool.Spec.ResourceShape.Labels[ResourceShapeLabelMachineTypes]
	if !ok {
		return []string{}
	}
	return xstring.SplitByCommaAndTrim(value)
}
