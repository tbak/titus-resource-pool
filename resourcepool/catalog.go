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
)
