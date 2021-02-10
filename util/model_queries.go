package util

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	v12 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

type NodeAndPods struct {
	Node *v1.Node
	Pods []*v1.Pod
}

func HasLabelAndValue(labels map[string]string, labelName string, value string) bool {
	if actual, ok := FindLabel(labels, labelName); ok {
		return value == actual
	}
	return false
}

// FIXME Resources are in MB but should be in bytes
func FromResourceListToComputeResource(limits v1.ResourceList) v12.ComputeResource {
	result := v12.ComputeResource{
		CPU:      limits.Cpu().Value(),
		MemoryMB: limits.Memory().Value() / OneMegaByte,
		DiskMB:   limits.StorageEphemeral().Value() / OneMegaByte,
	}
	if gpu, ok := limits[ResourceGpu]; ok {
		result.GPU += gpu.Value()
	}
	if network, ok := limits[ResourceNetwork]; ok {
		result.NetworkMBPS += network.Value() / OneMBPS
	}
	return result
}

func FromComputeResourceToResourceList(resources v12.ComputeResource) v1.ResourceList {
	return v1.ResourceList{
		v1.ResourceCPU:              resource.MustParse(strconv.FormatInt(resources.CPU, 10)),
		v1.ResourceMemory:           resource.MustParse(fmt.Sprintf("%vMi", resources.MemoryMB)),
		v1.ResourceEphemeralStorage: resource.MustParse(fmt.Sprintf("%vMi", resources.DiskMB)),
		ResourceGpu:                 resource.MustParse(strconv.FormatInt(resources.GPU, 10)),
		ResourceNetwork:             resource.MustParse(fmt.Sprintf("%vM", resources.NetworkMBPS)),
	}
}

func FindLabel(labels map[string]string, labelName string) (string, bool) {
	if labels == nil {
		return "", false
	}
	if actual, ok := labels[labelName]; ok {
		return actual, true
	}
	return "", false
}
