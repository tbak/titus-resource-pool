package pod

import (
	"strings"

	k8sCore "k8s.io/api/core/v1"

	commonNode "github.com/Netflix/titus-kube-common/node"
	commonPod "github.com/Netflix/titus-kube-common/pod"
)

func ButPodName(pod *k8sCore.Pod, name string) *k8sCore.Pod {
	pod.Name = name
	return pod
}

func ButPodLabel(pod *k8sCore.Pod, key string, value string) *k8sCore.Pod {
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	pod.Labels[key] = value
	return pod
}

func ButPodAnnotation(pod *k8sCore.Pod, key string, value string) *k8sCore.Pod {
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations[key] = value
	return pod
}

func ButPodMachineRequiredAffinity(pod *k8sCore.Pod, machineTypes []string) *k8sCore.Pod {
	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &k8sCore.Affinity{}
	}
	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &k8sCore.NodeAffinity{}
	}
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &k8sCore.NodeSelector{}
	}
	nodeSelector := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution

	term := k8sCore.NodeSelectorTerm{
		MatchExpressions: []k8sCore.NodeSelectorRequirement{
			k8sCore.NodeSelectorRequirement{
				Key:      commonNode.LabelKeyInstanceType,
				Operator: "In",
				Values:   machineTypes,
			},
		},
	}
	nodeSelector.NodeSelectorTerms = append(nodeSelector.NodeSelectorTerms, term)

	return pod
}

func ButPodResourcePools(pod *k8sCore.Pod, resourcePools ...string) *k8sCore.Pod {
	return ButPodLabel(pod, commonNode.LabelKeyResourcePool, strings.Join(resourcePools, ","))
}

func ButPodCapacityGroup(pod *k8sCore.Pod, capacityGroup string) *k8sCore.Pod {
	return ButPodLabel(pod, commonPod.LabelKeyCapacityGroup, capacityGroup)
}

func ButPodAssignedToNode(pod *k8sCore.Pod, node *k8sCore.Node) *k8sCore.Pod {
	pod.Spec.NodeName = node.Name
	return pod
}

func ButPodRunningOnNode(pod *k8sCore.Pod, node *k8sCore.Node) *k8sCore.Pod {
	pod = ButPodAssignedToNode(pod, node)
	pod.Status.Phase = k8sCore.PodRunning
	return pod
}
