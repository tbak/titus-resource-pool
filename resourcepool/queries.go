package resourcepool

import (
	"time"

	coreV1 "k8s.io/api/core/v1"

	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
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

// For a given resource pool:
// 1. find its all nodes and pods
// 2. map pods to their nodes
// 3. collect pods not running on any node in a separate list
func GroupNodesAndPods(resourcePool *poolV1.ResourcePoolSpec, allPods []*coreV1.Pod,
	allNodes []*coreV1.Node) (map[string]poolUtil.NodeAndPods, []*coreV1.Pod) {
	var nodesAndPodsMap = map[string]poolUtil.NodeAndPods{}
	var podsWithoutNode []*coreV1.Pod

	for _, node := range allNodes {
		if poolNode.NodeBelongsToResourcePool(node, resourcePool.Name) {
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
