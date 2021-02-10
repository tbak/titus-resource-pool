package resourcepool

import (
	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewResourcePoolCrdOf(name string, shapeDimensions v1.ComputeResource, shapeCount int64) *v1.ResourcePoolConfig {
	return &v1.ResourcePoolConfig{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1.ResourcePoolSpec{
			Name: name,
			ResourceShape: v1.ResourceShape{
				ComputeResource: shapeDimensions,
			},
			ScalingRules: v1.ResourcePoolScalingRules{
				MinIdle:            0,
				MaxIdle:            2,
				MinSize:            0,
				MaxSize:            10,
				AutoScalingEnabled: true,
			},
			ResourceCount: shapeCount,
			Status:        v1.ResourceDemandStatus{},
		},
	}
}

func NewResourcePoolCrdOfMachine(name string, machineTypeConfig *v1.MachineTypeConfig, partsCount int64,
	shapeCount int64) *v1.ResourcePoolConfig {
	shapeDimensions := machineTypeConfig.Spec.ComputeResource.Divide(partsCount)
	return NewResourcePoolCrdOf(name, shapeDimensions, shapeCount)
}
