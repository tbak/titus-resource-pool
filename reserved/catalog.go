package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-resource-pool/resourcepool"
	"github.com/Netflix/titus-resource-pool/util"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CapacityGroupNameEmpty = "emptyCapacityGroup"
)

// We use functions, as K8S records are mutable
var (
	EmptyCapacityGroup = func() *capacityGroupV1.CapacityGroup {
		return &capacityGroupV1.CapacityGroup{
			ObjectMeta: metaV1.ObjectMeta{
				Name: CapacityGroupNameEmpty,
			},
			Spec: capacityGroupV1.CapacityGroupSpec{
				CapacityGroupName: CapacityGroupNameEmpty,
				ResourcePoolName:  resourcepool.PoolNameIntegration,
				SchedulerName:     PodSchedulerKube,
				ComputeResource:   util.ComputeResourcesRegular.Multiply(8),
				InstanceCount:     0,
			},
		}
	}

	BasicCapacityGroup = func(name string, resourcePoolName string, shape v1.ComputeResource, count uint32) *capacityGroupV1.CapacityGroup {
		return &capacityGroupV1.CapacityGroup{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Spec: capacityGroupV1.CapacityGroupSpec{
				CapacityGroupName: name,
				ResourcePoolName:  resourcePoolName,
				SchedulerName:     PodSchedulerKube,
				ComputeResource:   shape,
				InstanceCount:     count,
			},
		}
	}
)
