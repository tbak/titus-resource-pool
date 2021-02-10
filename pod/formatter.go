package pod

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	v1 "k8s.io/api/core/v1"

	. "github.com/Netflix/titus-resource-pool/util"
)

func FormatPod(pod *v1.Pod, options FormatterOptions) string {
	if options.Level == FormatCompact {
		return formatPodCompact(pod)
	} else if options.Level == FormatEssentials {
		return formatPodEssentials(pod)
	} else if options.Level == FormatDetails {
		return ToJSONString(pod)
	}
	return formatPodCompact(pod)
}

func formatPodCompact(pod *v1.Pod) string {
	type Compact struct {
		Name  string
		State string
		Node  string
	}
	value := Compact{
		Name:  pod.Name,
		State: toPodState(pod),
		Node:  pod.Spec.NodeName,
	}
	return ToJSONString(value)
}

func formatPodEssentials(pod *v1.Pod) string {
	type Compact struct {
		Name             string
		State            string
		Node             string
		ComputeResources poolV1.ComputeResource
	}
	value := Compact{
		Name:             pod.Name,
		State:            toPodState(pod),
		Node:             pod.Spec.NodeName,
		ComputeResources: FromPodToComputeResource(pod),
	}
	return ToJSONString(value)
}

func toPodState(pod *v1.Pod) string {
	if IsPodRunning(pod) {
		return "running"
	}
	if IsPodFinished(pod) {
		return "finished"
	}
	return "notScheduled"
}
