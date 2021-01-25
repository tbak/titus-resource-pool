package pod

import (
	k8sCore "k8s.io/api/core/v1"

	commonNode "github.com/Netflix/titus-kube-common/node"
)

func ButPodName(pod *k8sCore.Pod, name string) *k8sCore.Pod {
	pod.Name = name
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
