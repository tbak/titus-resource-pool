package machine

import (
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

// We use functions, as K8S records are mutable
var (
	// In certain cases it is important to not make any changes if a resource size update is less than a machine size.
	// Rounding errors may cause oscillations (like infinite scale ups and downs). `TheBiggestMachineThatCouldBe` should
	// be bigger in every dimension than every machine type used. It cannot be too big however, as this may have other
	// side effects, like too conservative decisions slowing down scale up or down operations.
	TheBiggestMachineThatCouldBe = func() *machineTypeV1.MachineTypeConfig {
		return &machineTypeV1.MachineTypeConfig{
			ObjectMeta: v12.ObjectMeta{
				Name:      "theBigOne",
				Namespace: "default",
			},
			Spec: machineTypeV1.MachineType{
				Name: "theBigOne",
				ComputeResource: poolV1.ComputeResource{
					CPU:         96,
					MemoryMB:    800000,
					DiskMB:      2000000,
					NetworkMBPS: 25000,
				},
			},
		}
	}
	TheBiggestMachineThatCouldBeResources = TheBiggestMachineThatCouldBe().Spec.ComputeResource

	M5Metal = func() *machineTypeV1.MachineTypeConfig {
		return &machineTypeV1.MachineTypeConfig{
			ObjectMeta: v12.ObjectMeta{
				Name:      "m5.metal",
				Namespace: "default",
			},
			Spec: machineTypeV1.MachineType{
				Name: "m5.metal",
				ComputeResource: poolV1.ComputeResource{
					CPU:         96,
					MemoryMB:    393216,
					DiskMB:      1048576,
					NetworkMBPS: 25000,
				},
			},
		}
	}
	R5Metal = func() *machineTypeV1.MachineTypeConfig {
		return &machineTypeV1.MachineTypeConfig{
			ObjectMeta: v12.ObjectMeta{
				Name:      "r5.metal",
				Namespace: "default",
			},
			Spec: machineTypeV1.MachineType{
				Name: "r5.metal",
				ComputeResource: poolV1.ComputeResource{
					CPU:         96,
					MemoryMB:    786432,
					DiskMB:      1536000,
					NetworkMBPS: 25000,
				},
			},
		}
	}
)
