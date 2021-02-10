package pod

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"github.com/Netflix/titus-kube-common/node"
	"github.com/Netflix/titus-resource-pool/machine"
	node2 "github.com/Netflix/titus-resource-pool/node"
	. "github.com/Netflix/titus-resource-pool/util"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNotScheduledPodWithName(name string, resourcePoolName string, resources v1.ComputeResource,
	now time.Time) *v13.Pod {
	return &v13.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: "default",
			CreationTimestamp: v12.Time{
				Time: now,
			},
			Labels: map[string]string{
				node.LabelKeyResourcePool: resourcePoolName,
			},
		},
		Spec: v13.PodSpec{
			Containers: []v13.Container{
				{
					Name:  "main",
					Image: "some/image:latest",
					Resources: v13.ResourceRequirements{
						Limits:   FromComputeResourceToResourceList(resources),
						Requests: FromComputeResourceToResourceList(resources),
					},
				},
			},
		},
		Status: v13.PodStatus{},
	}
}

func NewNotScheduledPod(resourcePoolName string, resources v1.ComputeResource, now time.Time) *v13.Pod {
	return NewNotScheduledPodWithName(uuid.New().String()+".pod", resourcePoolName, resources, now)
}

func NewRandomNotScheduledPod() *v13.Pod {
	return NewNotScheduledPod(node2.ResourcePoolElastic, machine.R5Metal().Spec.ComputeResource.Divide(4), time.Now())
}

func NewNotScheduledPods(count int64, namePrefix string, resourcePoolName string, resources v1.ComputeResource,
	now time.Time) []*v13.Pod {
	var pods []*v13.Pod
	for i := int64(0); i < count; i++ {
		pods = append(pods, NewNotScheduledPodWithName(fmt.Sprintf("%v#%v", namePrefix, i), resourcePoolName,
			resources, now))
	}
	return pods
}
