package resourcepool

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PoolNameEmpty        = "emptyResourcePool"
	PoolNameIntegration  = "integrationTestResourcePool"
	PoolNameIntegration2 = "integrationTestResourcePool2"
)

// We use functions, as K8S records are mutable
var (
	EmptyResourcePool = func() *poolV1.ResourcePoolConfig {
		return &poolV1.ResourcePoolConfig{
			ObjectMeta: metaV1.ObjectMeta{
				Name: PoolNameEmpty,
			},
			Spec: poolV1.ResourcePoolSpec{
				Name: PoolNameEmpty,
			},
		}
	}

	BasicResourcePool = func(name string, instanceCount int64, shape poolV1.ComputeResource) *poolV1.ResourcePoolConfig {
		pool := ButResourcePoolName(EmptyResourcePool(), name)
		pool.Spec.ResourceShape = poolV1.ResourceShape{
			ComputeResource: shape,
			Labels:          map[string]string{},
		}
		pool.Spec.ResourceCount = instanceCount
		return pool
	}
)
