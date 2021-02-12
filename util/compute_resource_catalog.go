package util

import v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"

// Collection of constants and functions to generate compute resources of different sizes.

var (
	ComputeResourceLowCPU = v1.ComputeResource{
		CPU:         1,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}
	ComputeResourcesRegular = v1.ComputeResource{
		CPU:         4,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}
	ComputeResourcesHighCPU = v1.ComputeResource{
		CPU:         20,
		GPU:         0,
		MemoryMB:    8096,
		DiskMB:      16384,
		NetworkMBPS: 256,
	}
)
