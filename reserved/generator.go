package reserved

import (
	"github.com/Netflix/titus-resource-pool/resourcepool"
	"github.com/google/uuid"

	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

func NewCapacityGroup(name string, resourcePoolName string) *v1.CapacityGroup {
	result := EmptyCapacityGroup()
	result.ObjectMeta.Name = name
	result.Spec.CapacityGroupName = name
	result.Spec.ResourcePoolName = resourcePoolName
	result.Spec.InstanceCount = 5
	return result
}

func NewRandomCapacityGroup(transformers ...func(node *v1.CapacityGroup)) *v1.CapacityGroup {
	node := NewCapacityGroup(uuid.New().String()+".capacityGroup", resourcepool.PoolNameIntegration)
	for _, transformer := range transformers {
		transformer(node)
	}
	return node
}
