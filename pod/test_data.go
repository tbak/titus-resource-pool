package pod

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	poolApi "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	commonNode "github.com/Netflix/titus-kube-common/node"
	poolMachine "github.com/Netflix/titus-resource-pool/machine"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func NewNotScheduledPodWithName(name string, resourcePoolName string, resources poolApi.ComputeResource,
	now time.Time) *coreV1.Pod {
	return &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			CreationTimestamp: metaV1.Time{
				Time: now,
			},
			Labels: map[string]string{
				commonNode.LabelKeyResourcePool: resourcePoolName,
			},
		},
		Spec: coreV1.PodSpec{
			Containers: []coreV1.Container{
				{
					Name:  "main",
					Image: "some/image:latest",
					Resources: coreV1.ResourceRequirements{
						Limits:   poolUtil.FromComputeResourceToResourceList(resources),
						Requests: poolUtil.FromComputeResourceToResourceList(resources),
					},
				},
			},
		},
		Status: coreV1.PodStatus{},
	}
}

func NewNotScheduledPod(resourcePoolName string, resources poolApi.ComputeResource, now time.Time) *coreV1.Pod {
	return NewNotScheduledPodWithName(uuid.New().String()+".pod", resourcePoolName, resources, now)
}

func NewRandomNotScheduledPod() *coreV1.Pod {
	return NewNotScheduledPod(poolNode.ResourcePoolElastic,
		poolMachine.R5Metal().Spec.ComputeResource.Divide(4), time.Now())
}

func NewNotScheduledPods(count int64, namePrefix string, resourcePoolName string, resources poolApi.ComputeResource,
	now time.Time) []*coreV1.Pod {
	var pods []*coreV1.Pod
	for i := int64(0); i < count; i++ {
		pods = append(pods, NewNotScheduledPodWithName(fmt.Sprintf("%v#%v", namePrefix, i), resourcePoolName,
			resources, now))
	}
	return pods
}
