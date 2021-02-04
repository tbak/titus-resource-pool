package resourcepool

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"strings"
)

func ButResourcePoolName(pool *poolV1.ResourcePoolConfig, name string) *poolV1.ResourcePoolConfig {
	pool.Name = name
	pool.Spec.Name = name
	return pool
}

func ButResourceShapeLabels(pool *poolV1.ResourcePoolConfig, pairs ...string) *poolV1.ResourcePoolConfig {
	labels := pool.Spec.ResourceShape.Labels
	if labels == nil {
		pool.Spec.ResourceShape.Labels = map[string]string{}
		labels = pool.Spec.ResourceShape.Labels
	}
	steps := (len(pairs) + 1) / 2
	for i := 0; i < steps; i++ {
		pos := i * 2
		key := pairs[pos]
		if pos+1 < len(pairs) {
			labels[key] = pairs[pos+1]
		} else {
			labels[key] = ""
		}
	}
	return pool
}

func ButResourcePoolMachineTypes(pool *poolV1.ResourcePoolConfig, machineTypes []string) *poolV1.ResourcePoolConfig {
	if len(machineTypes) == 0 {
		return pool
	}
	return ButResourceShapeLabels(pool, ResourceShapeLabelMachineTypes, strings.Join(machineTypes, ","))
}
