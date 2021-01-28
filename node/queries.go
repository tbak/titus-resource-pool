package node

import (
	"time"

	k8sCore "k8s.io/api/core/v1"

	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func AsNodeReferenceList(nodeList *k8sCore.NodeList) []*k8sCore.Node {
	result := []*k8sCore.Node{}
	for _, node := range nodeList.Items {
		tmp := node
		result = append(result, &tmp)
	}
	return result
}

// Resolve node state.
func UniqueNodeState(node *k8sCore.Node, now time.Time, ageThreshold time.Duration) string {
	if poolUtil.IsNodeBootstrapping(node, now, ageThreshold) {
		return NodeStateBootstrapping
	}
	if poolUtil.IsNodeAvailableForScheduling(node, now, ageThreshold) {
		return NodeStateActive
	}
	if poolUtil.IsNodeDecommissioned(node) {
		return NodeStateDecommissioned
	}
	if poolUtil.IsNodeScalingDown(node) {
		return NodeStateScalingDown
	}
	if poolUtil.IsNodeRemovable(node) {
		return NodeStateRemovable
	}
	return NodeStateBroken
}
