package pod

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	v1 "k8s.io/api/core/v1"

	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func FormatPod(pod *v1.Pod, options poolUtil.FormatterOptions) string {
	if options.Level == poolUtil.FormatCompact {
		return formatPodCompact(pod)
	} else if options.Level == poolUtil.FormatEssentials {
		return formatPodEssentials(pod)
	} else if options.Level == poolUtil.FormatDetails {
		return poolUtil.ToJSONString(pod)
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
	return poolUtil.ToJSONString(value)
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
	return poolUtil.ToJSONString(value)
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
