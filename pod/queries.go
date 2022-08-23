package pod

import (
	"strings"
	"time"

	v1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"

	k8sCore "k8s.io/api/core/v1"

	poolApi "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	commonNode "github.com/Netflix/titus-kube-common/node"
	commonPod "github.com/Netflix/titus-kube-common/pod"
	poolNode "github.com/Netflix/titus-resource-pool/node"
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
	var assigned string
	var ok bool
	if assigned, ok = poolUtil.FindLabel(pod.Labels, commonPod.LabelKeyCapacityGroup); !ok {
		assigned, _ = poolUtil.FindLabel(pod.Annotations, commonPod.LabelKeyCapacityGroup)
	}
	return assigned
}

func IsPodInCapacityGroup(pod *k8sCore.Pod, cg *v1.CapacityGroup) bool {
	return FindPodCapacityGroup(pod) == cg.Spec.OriginalName
}

func IsPodWaitingToBeScheduled(pod *k8sCore.Pod) bool {
	// in some rare cases, a pod killed before scheduling might stay in Pending phase without
	// a nodeName but with a DeletionTimestamp
	return pod.Spec.NodeName == "" && !IsPodFinished(pod) && pod.ObjectMeta.DeletionTimestamp == nil
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

// A pod may be assigned to multiple resource pools. The first one returned is considered the primary which will
// be scaled up if more capacity is needed.
func FindPodAssignedResourcePools(pod *k8sCore.Pod) ([]string, bool) {
	var poolNames string
	var ok bool
	if poolNames, ok = poolUtil.FindLabel(pod.Labels, commonNode.LabelKeyResourcePool); !ok {
		if poolNames, ok = poolUtil.FindLabel(pod.Annotations, commonNode.LabelKeyResourcePool); !ok {
			return []string{}, false
		}
	}
	if poolNames == "" {
		return []string{}, false
	}
	parts := strings.Split(poolNames, ",")
	var names []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) > 0 {
			names = append(names, trimmed)
		}
	}
	if len(names) == 0 {
		return []string{}, false
	}
	return names, true
}

func FindPodPrimaryResourcePool(pod *k8sCore.Pod) (string, bool) {
	if poolNames, ok := FindPodAssignedResourcePools(pod); ok {
		return poolNames[0], true
	}
	return "", false
}

func IsPodInPrimaryResourcePool(resourcePool string, pod *k8sCore.Pod) bool {
	if primary, ok := FindPodPrimaryResourcePool(pod); ok {
		if primary == resourcePool {
			return true
		}
	}
	return false
}

// Find all pods for which the given resource pool is primary.
func FindPodsWithPrimaryResourcePool(resourcePool string, pods []*k8sCore.Pod) []*k8sCore.Pod {
	var result []*k8sCore.Pod
	for _, pod := range pods {
		if IsPodInPrimaryResourcePool(resourcePool, pod) {
			result = append(result, pod)
		}
	}
	return result
}

// TODO remove
func PodBelongsToResourcePool(pod *k8sCore.Pod, assignedPools []string, resourcePool string,
	resourcePoolWithGpus bool, nodes map[string]*k8sCore.Node) bool {
	if len(assignedPools) == 0 {
		return false
	}

	// Do not look at pods requesting GPU resources, but running in non-GPU resource pool.
	if !resourcePoolWithGpus {
		for _, container := range pod.Spec.Containers {
			if poolUtil.GetGpu(container.Resources.Requests) > 0 {
				return false
			}
		}
	}

	for _, pool := range assignedPools {
		if pool == resourcePool {
			// If the pod is not assigned to any node, we stop at this point.
			if pod.Spec.NodeName == "" {
				return true
			}
			// If the pod is assigned to a node, we check that the node itself belongs to the same resource pool.
			if node, ok := nodes[pod.Spec.NodeName]; ok {
				if poolNode.NodeBelongsToResourcePool(node, resourcePool) {
					return true
				}
			}
			return false
		}
	}

	return false
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
