package resourcepool

import (
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"

	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	commonNode "github.com/Netflix/titus-kube-common/node"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
	"github.com/Netflix/titus-resource-pool/util/xstring"
)

// Machine types used by a resource pool or empty array if none is defined.
func GetResourcePoolMachineTypes(resourcePool *poolV1.ResourcePoolConfig) []string {
	value, ok := resourcePool.Spec.ResourceShape.Labels[ResourceShapeLabelMachineTypes]
	if !ok {
		return []string{}
	}
	return xstring.SplitByCommaAndTrim(value)
}

func PodBelongsToResourcePool(pod *coreV1.Pod, resourcePool *poolV1.ResourcePoolSpec, nodes []*coreV1.Node) bool {
	// Do not look at pods requesting GPU resources, but running in non-GPU resource pool.
	if resourcePool.ResourceShape.GPU <= 0 {
		for _, container := range pod.Spec.Containers {
			if poolUtil.FromResourceListToComputeResource(container.Resources.Requests).GPU > 0 {
				return false
			}
		}
	}
	assignedPools, ok := FindPodAssignedResourcePools(pod)
	if !ok {
		return false
	}

	for _, pool := range assignedPools {
		if pool == resourcePool.Name {
			// If the pod is not assigned to any node, we stop at this point.
			if pod.Spec.NodeName == "" {
				return true
			}
			// If the pod is assigned to a node, we check that the node itself belongs to the same resource pool.
			for _, node := range nodes {
				if NodeBelongsToResourcePool(node, resourcePool) && node.Name == pod.Spec.NodeName {
					return true
				}
			}
			return false
		}
	}

	return false
}

func NodeBelongsToResourcePool(node *coreV1.Node, resourcePool *poolV1.ResourcePoolSpec) bool {
	return poolUtil.HasLabelAndValue(node.Labels, commonNode.LabelKeyResourcePool, resourcePool.Name)
}

// A pod may be assigned to multiple resource pools. The first one returned is considered the primary which will
// be scaled up if more capacity is needed.
func FindPodAssignedResourcePools(pod *coreV1.Pod) ([]string, bool) {
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

func FindPodPrimaryResourcePool(pod *coreV1.Pod) (string, bool) {
	if poolNames, ok := FindPodAssignedResourcePools(pod); ok {
		return poolNames[0], true
	}
	return "", false
}

// Find all pods for which the given resource pool is primary.
func FindPodsWithPrimaryResourcePool(resourcePool string, pods []*coreV1.Pod) []*coreV1.Pod {
	var result []*coreV1.Pod
	for _, pod := range pods {
		if primary, ok := FindPodPrimaryResourcePool(pod); ok {
			if primary == resourcePool {
				result = append(result, pod)
			}
		}
	}
	return result
}

// For a given resource pool:
// 1. find its all nodes and pods
// 2. map pods to their nodes
// 3. collect pods not running on any node in a separate list
func GroupNodesAndPods(resourcePool *poolV1.ResourcePoolSpec, allPods []*coreV1.Pod,
	allNodes []*coreV1.Node) (map[string]poolUtil.NodeAndPods, []*coreV1.Pod) {
	var nodesAndPodsMap = map[string]poolUtil.NodeAndPods{}
	var podsWithoutNode []*coreV1.Pod

	for _, node := range allNodes {
		if NodeBelongsToResourcePool(node, resourcePool) {
			nodesAndPodsMap[node.Name] = poolUtil.NodeAndPods{Node: node}
		}
	}
	for _, pod := range allPods {
		// Do not include finished pods
		if pod.Status.Phase != "Succeeded" && pod.Status.Phase != "Failed" {
			// We do not check if a pod is directly associated with the resource pool, as we only care
			// that it runs on a node that belongs to it.
			if nodeAndPods, ok := nodesAndPodsMap[pod.Spec.NodeName]; ok {
				nodeAndPods.Pods = append(nodeAndPods.Pods, pod)
				nodesAndPodsMap[pod.Spec.NodeName] = nodeAndPods
			} else {
				podsWithoutNode = append(podsWithoutNode, pod)
			}
		}
	}
	return nodesAndPodsMap, podsWithoutNode
}

func GroupNodesByLifecycleState(nodes []*coreV1.Node, now time.Time,
	nodeBootstrapThreshold time.Duration) ([]*coreV1.Node, []*coreV1.Node, []*coreV1.Node) {
	comingUp := []*coreV1.Node{}
	schedulable := []*coreV1.Node{}
	comingDown := []*coreV1.Node{}
	for _, node := range nodes {
		if poolNode.IsNodeBootstrapping(node, now, nodeBootstrapThreshold) {
			comingUp = append(comingUp, node)
		} else if poolNode.IsNodeOnItsWayOut(node) {
			comingDown = append(comingDown, node)
		} else {
			schedulable = append(schedulable, node)
		}
	}
	return comingUp, schedulable, comingDown
}
