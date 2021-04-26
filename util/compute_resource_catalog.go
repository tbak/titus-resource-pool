package util

import (
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	titusPool "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

// Collection of constants and functions to generate compute resources of different sizes.

var (
	// Deprecated
	ComputeResourceLowCPU = titusPool.ComputeResource{
		CPU:         1,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}
	ComputeResourcesRegular = titusPool.ComputeResource{
		CPU:         4,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}
	ComputeResourcesHighCPU = titusPool.ComputeResource{
		CPU:         20,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}

	// Compute resource elementary units used in the integration tests. All actual resource dimensions (pools, nodes,
	// pods, etc) are defined as multiplication of those here.

	// Proportional resoure allocation accross all dimensions (excluding GPU which is a special case).
	ComputeResourcesUnitProportional = titusPool.ComputeResource{
		CPU:         1,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}

	// Test machine types. The variable names encodes the compute resource unit utilized and its CPU multiplication factor.

	MachineFromUnitProportional96 = func() *machineTypeV1.MachineTypeConfig {
		return &machineTypeV1.MachineTypeConfig{
			ObjectMeta: v12.ObjectMeta{
				Name:      "test.proportional96",
				Namespace: "default",
			},
			Spec: machineTypeV1.MachineType{
				Name:            "m5.metal",
				ComputeResource: ComputeResourcesUnitProportional.Multiply(96),
			},
		}
	}
)
