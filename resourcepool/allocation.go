package resourcepool

import (
	v1 "k8s.io/api/core/v1"

	scaler "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"

	poolNode "github.com/Netflix/titus-resource-pool/node"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
)

func ComputeAllocatableCapacityFromSnapshot(snapshot *ResourceSnapshot,
	minimumResources scaler.ComputeResource, adjust bool) (
	scaler.ComputeResource, scaler.ComputeResource, map[string]scaler.ComputeResource) {
	scheduledPods := snapshot.PodSnapshot.ScheduledByName
	return ComputeAllocatableCapacity(scheduledPods, snapshot.NodeSnapshot.ActiveByName, minimumResources, adjust)
}

// In order to compute allocatable capacity and be robust to the opportunistic-cpu-driven oversubscription case,
// we cannot work in aggregate and subtract total allocated capacity from total provisioned capacity.
// We need to do the accounting per node, and then sum.
// Second return value is actual available capacity per node without taking any DRF adjustment into account.
// It gives an upper bound on the available capacity that may be used to evaluate reservation shortage.
// The third return value is a map that contains actual remaining capacity per node without taking any
// DRF resource adjustment and minimum resource size check for debugging purposes.
func ComputeAllocatableCapacity(scheduledPods map[string]*v1.Pod, nodes map[string]*v1.Node,
	minimumResources scaler.ComputeResource, adjust bool) (
	scaler.ComputeResource, scaler.ComputeResource, map[string]scaler.ComputeResource) {
	nodeToAvailable := make(map[string]scaler.ComputeResource)
	nodeToUsed := make(map[string]scaler.ComputeResource)

	// Total running nodes capacity
	for _, node := range nodes {
		nodeToAvailable[node.Name] = poolNode.FromNodeToComputeResource(node)
		nodeToUsed[node.Name] = scaler.ComputeResource{}
	}

	// Used capacity per node. We only look at pods running on the active nodes.
	for _, pod := range scheduledPods {
		nodeName := pod.Spec.NodeName
		if nodeUsed, exists := nodeToUsed[nodeName]; exists {
			nodeToUsed[nodeName] = nodeUsed.Add(poolPod.FromPodToComputeResource(pod))
		}
	}

	// Sum what is remaining, but only look at nodes with large enough resource chunks left.
	// Align the remaining resources to the mostly utilized resource.
	tot := scaler.ComputeResource{}
	remainingActual := scaler.ComputeResource{}
	nodeRemainingCapacityDebug := map[string]scaler.ComputeResource{}
	for nodeID, nodeUsed := range nodeToUsed {
		nodeAvailable := nodeToAvailable[nodeID]
		nodeRemaining := nodeAvailable.SubWithLimit(nodeUsed, 0)
		if nodeRemaining.GreaterThanOrEqual(minimumResources) {
			adjustedUsed := nodeUsed
			if adjust {
				adjustedUsed = nodeUsed.AlignResourceRatios(nodeAvailable)
			}
			remainingAdjusted := nodeAvailable.SubWithLimit(adjustedUsed, 0)
			tot = tot.Add(remainingAdjusted)
			remainingActual = remainingActual.Add(nodeRemaining)
		}
		nodeRemainingCapacityDebug[nodeID] = nodeRemaining
	}

	return tot, remainingActual, nodeRemainingCapacityDebug
}
