package pod

import (
	"github.com/Netflix/titus-resource-pool/util/xcollection"
	k8sCore "k8s.io/api/core/v1"

	commonNode "github.com/Netflix/titus-kube-common/node"
)

// TODO Remove when no longer in use
const legacyInstanceTypeLabel = "beta.kubernetes.io/instance-type"

func AsPodReferenceList(podList *k8sCore.PodList) []*k8sCore.Pod {
	result := []*k8sCore.Pod{}
	for _, node := range podList.Items {
		tmp := node
		result = append(result, &tmp)
	}
	return result
}

// Return machine types explicitly requested by pod using hard affinity rules.
func GetPodRequestedMachineTypes(pod *k8sCore.Pod) []string {
	if pod.Spec.Affinity == nil || pod.Spec.Affinity.NodeAffinity == nil ||
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return []string{}
	}
	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	for _, term := range terms {
		for _, expr := range term.MatchExpressions {
			if expr.Key == commonNode.LabelKeyInstanceType || expr.Key == legacyInstanceTypeLabel {
				return expr.Values
			}
		}
	}
	return []string{}
}

// See IsPodOkWithMachineTypes to understand the filtering criteria.
func FilterPodsOkWithMachineTypes(pods []*k8sCore.Pod, machineTypes []string) []*k8sCore.Pod {
	var result []*k8sCore.Pod
	machineSet := xcollection.SetOfStringList(machineTypes)
	for _, pod := range pods {
		if IsPodOkWithMachineTypesSet(pod, machineSet) {
			result = append(result, pod)
		}
	}
	return result
}

// Returns true if the given pod can be run on the provided machine types. If the machine types list is empty, returns
// false. Otherwise for a pod to not match, it must have a machine hard constraint with machine types disjoned with
// the provided set.
//
// For example if machinesTypes=["r5.metal", "m5.metal"] and pod requires ["c5.metal], it will not be added to the
// result.
// If machinesTypes=["r5.metal", "m5.metal"] and pod requires ["r5.metal", "c5.metal], it will be added to the result.
func IsPodOkWithMachineTypesSet(pod *k8sCore.Pod, machineTypes map[string]bool) bool {
	if len(machineTypes) == 0 {
		return false
	}
	requestedList := GetPodRequestedMachineTypes(pod)
	if len(requestedList) == 0 {
		return true
	}
	for _, next := range requestedList {
		if _, ok := machineTypes[next]; ok {
			return true
		}
	}
	return false
}
