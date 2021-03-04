package pod

import (
	"strings"
	"time"

	k8sCore "k8s.io/api/core/v1"

	poolApi "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	commonNode "github.com/Netflix/titus-kube-common/node"
	commonPod "github.com/Netflix/titus-kube-common/pod"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
	"github.com/Netflix/titus-resource-pool/util/xcollection"
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

func FindPodCapacityGroup(pod *k8sCore.Pod) string {
	if assigned, ok := poolUtil.FindLabel(pod.Labels, commonPod.LabelKeyCapacityGroup); ok {
		return assigned
	}
	assigned, _ := poolUtil.FindLabel(pod.Annotations, commonPod.LabelKeyCapacityGroup)
	return assigned
}

func IsPodInCapacityGroup(pod *k8sCore.Pod, capacityGroupName string) bool {
	podCapacityGroup := FindPodCapacityGroup(pod)
	return podCapacityGroup == capacityGroupName || strings.ReplaceAll(podCapacityGroup, "_", "-") == capacityGroupName
}

func IsPodWaitingToBeScheduled(pod *k8sCore.Pod) bool {
	return pod.Spec.NodeName == "" && !IsPodFinished(pod)
}

func IsPodRunning(pod *k8sCore.Pod) bool {
	if IsPodFinished(pod) {
		return false
	}
	return pod.Spec.NodeName != ""
}

func IsPodFinished(pod *k8sCore.Pod) bool {
	return pod.Status.Phase == k8sCore.PodSucceeded || pod.Status.Phase == k8sCore.PodFailed
}

func Age(pod *k8sCore.Pod, now time.Time) time.Duration {
	return now.Sub(pod.CreationTimestamp.Time)
}

func FromPodToComputeResource(pod *k8sCore.Pod) poolApi.ComputeResource {
	total := poolApi.ComputeResource{}
	for _, container := range pod.Spec.Containers {
		total = total.Add(poolUtil.FromResourceListToComputeResource(container.Resources.Requests))
	}
	return total
}

func Names(pods *[]k8sCore.Pod) []string {
	var names []string
	for _, node := range *pods {
		names = append(names, node.Name)
	}
	return names
}

func FindNotScheduledPods(pods []*k8sCore.Pod) []*k8sCore.Pod {
	var waiting []*k8sCore.Pod
	for _, pod := range pods {
		if IsPodWaitingToBeScheduled(pod) {
			waiting = append(waiting, pod)
		}
	}
	return waiting
}

// Find all unscheduled pods belonging to the given resource pool, which are not younger than a threshold.
func FindOldNotScheduledPods(pods []*k8sCore.Pod, youngPodThreshold time.Duration, now time.Time) []*k8sCore.Pod {
	var waiting []*k8sCore.Pod
	for _, pod := range pods {
		if IsPodWaitingToBeScheduled(pod) && Age(pod, now) >= youngPodThreshold {
			waiting = append(waiting, pod)
		}
	}
	return waiting
}

func FilterRunningPods(pods []*k8sCore.Pod) []*k8sCore.Pod {
	var active []*k8sCore.Pod
	for _, pod := range pods {
		if IsPodRunning(pod) {
			active = append(active, pod)
		}
	}
	return active
}

func CountNotScheduledPods(pods []*k8sCore.Pod) int64 {
	var count int64
	for _, pod := range pods {
		if IsPodWaitingToBeScheduled(pod) {
			count = count + 1
		}
	}
	return count
}

func SumPodResources(pods []*k8sCore.Pod) poolApi.ComputeResource {
	var sum poolApi.ComputeResource
	for _, pod := range pods {
		sum = sum.Add(FromPodToComputeResource(pod))
	}
	return sum
}
