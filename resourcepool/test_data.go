package resourcepool

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

func NewResourcePoolCrdOf(name string, shapeDimensions poolV1.ComputeResource, shapeCount int64) *poolV1.ResourcePoolConfig {
	return &poolV1.ResourcePoolConfig{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: poolV1.ResourcePoolSpec{
			Name: name,
			ResourceShape: poolV1.ResourceShape{
				ComputeResource: shapeDimensions,
			},
			ScalingRules: poolV1.ResourcePoolScalingRules{
				MinIdle:            0,
				MaxIdle:            2,
				MinSize:            0,
				MaxSize:            10,
				AutoScalingEnabled: true,
			},
			ResourceCount: shapeCount,
			Status:        poolV1.ResourceDemandStatus{},
		},
	}
}

func NewResourcePoolCrdOfMachine(name string, machineTypeConfig *machineTypeV1.MachineTypeConfig, partsCount int64,
	shapeCount int64) *poolV1.ResourcePoolConfig {
	shapeDimensions := machineTypeConfig.Spec.ComputeResource.Divide(partsCount)
	return NewResourcePoolCrdOf(name, shapeDimensions, shapeCount)
}
