package reserved

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/resourcepool"
	"github.com/Netflix/titus-resource-pool/util"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CapacityGroupNameEmpty = "emptyCapacityGroup"
)

// We use functions, as K8S records are mutable
var (
	EmptyCapacityGroup = func() *poolV1.CapacityGroup {
		return &poolV1.CapacityGroup{
			ObjectMeta: metaV1.ObjectMeta{
				Name: CapacityGroupNameEmpty,
			},
			Spec: poolV1.CapacityGroupSpec{
				CapacityGroupName: CapacityGroupNameEmpty,
				ResourcePoolName:  resourcepool.PoolNameIntegration,
				SchedulerName:     PodSchedulerKube,
				ComputeResource:   util.ComputeResourcesRegular.Multiply(8),
				InstanceCount:     0,
			},
		}
	}
)
