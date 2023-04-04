package node

import (
	"sort"
	"time"

	k8sCore "k8s.io/api/core/v1"

	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	poolApi "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	commonNode "stash.corp.netflix.com/tn/titus-kube-common/node"
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
	if IsNodeBootstrapping(node, now, ageThreshold) {
		return NodeStateBootstrapping
	}
	if IsNodeAvailableForScheduling(node, now, ageThreshold) {
		return NodeStateActive
	}
	if IsNodeDecommissioned(node) {
		return NodeStateDecommissioned
	}
	if IsNodePhasedOut(node) {
		return NodeStatePhasedOut
	}
	if IsNodeScalingDown(node) {
		return NodeStateScalingDown
	}
	if IsNodeRemovable(node) {
		return NodeStateRemovable
	}
	return NodeStateBroken
}

// Returns true, if a node is a Kubelet node
func IsKubeletNode(node *k8sCore.Node) bool {
	if backend, ok := poolUtil.FindLabel(node.Labels, NodeLabelBackend); ok {
		return backend == NodeBackendKubelet
	}
	return false
}

func HasNoExecuteTaint(node *k8sCore.Node) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Effect == "NoExecute" {
			return true
		}
	}
	return false
}

// Returns true for a new node that is still bootstrapping. `ageThreshold` is a time limit for a node to be
// regarded as new.
func IsNodeBootstrapping(node *k8sCore.Node, now time.Time, ageThreshold time.Duration) bool {
	// This taint explicitly tells us that the node is initializing.
	if FindTaint(node, commonNode.TaintKeyInit) != nil {
		return true
	}

	if node.CreationTimestamp.Add(ageThreshold).Before(now) {
		return false
	}

	// Getting here does not guarantee (at least at the time of writing this change), that the new node is
	// fully initialized and ready to take traffic. We make here a few heuristic guesses to improve th evaluation
	// accuracy.
	return IsNodeBroken(node)
}

func IsNodeBootstrapping2(node *k8sCore.Node, pastDeadline func(*k8sCore.Node) bool) bool {
	// This taint explicitly tells us that the node is initializing.
	if FindTaint(node, commonNode.TaintKeyInit) != nil {
		return true
	}

	if pastDeadline(node) {
		return false
	}

	// Getting here does not guarantee (at least at the time of writing this change), that the new node is
	// fully initialized and ready to take traffic. We make here a few heuristic guesses to improve th evaluation
	// accuracy.
	return IsNodeBroken(node)
}

func IsNodeBroken(node *k8sCore.Node) bool {
	// FIXME Discern better between actual bad states and other cases
	if HasNoExecuteTaint(node) {
		return true
	}

	// It happens that there are node objects registered with no resources
	if node.Status.Allocatable.Cpu().IsZero() {
		return true
	}

	return false
}

func IsNodeAvailableForScheduling(node *k8sCore.Node, now time.Time, ageThreshold time.Duration) bool {
	return !IsNodeBootstrapping(node, now, ageThreshold) &&
		!IsNodeToRemove(node) &&
		!IsNodeRemovable(node) &&
		!IsNodeTerminated(node)
}

func IsNodeOnItsWayOut(node *k8sCore.Node) bool {
	return IsNodeToRemove(node) || IsNodeRemovable(node) || IsNodeTerminated(node)
}

// IsNodeDecommissioned returns true if the node has a decommissioning
// taint with NoExecute effect. Nodes with decommissioning taint with any other effects (NoSchedule
// and PreferNoSchedule) are considered schedulable.
func IsNodeDecommissioned(node *k8sCore.Node) bool {
	taint := FindTaint(node, commonNode.TaintKeyNodeDecommissioning)
	return taint != nil && taint.Effect == k8sCore.TaintEffectNoExecute
}

// IsNodePhasedOut returns true if the node is tainted decommissioning with NoSchedule or PreferNoSchedule effect.
// Such nodes are considered schedulable for most workloads with an exception of Titus Jobs with ActiveHost
// hard constraints.
func IsNodePhasedOut(node *k8sCore.Node) bool {
	taint := FindTaint(node, commonNode.TaintKeyNodeDecommissioning)
	return taint != nil && (taint.Effect == k8sCore.TaintEffectPreferNoSchedule ||
		taint.Effect == k8sCore.TaintEffectNoSchedule)
}

func IsNodeScalingDown(node *k8sCore.Node) bool {
	return FindTaint(node, commonNode.TaintKeyNodeScalingDown) != nil
}

func IsNodeEvacuating(node *k8sCore.Node) bool {
	taint := FindTaint(node, commonNode.TaintKeyNodeEvacuate)
	return taint != nil && taint.Effect == k8sCore.TaintEffectNoExecute
}

func IsNodeToRemove(node *k8sCore.Node) bool {
	return IsNodeDecommissioned(node) || IsNodeScalingDown(node) || IsNodeEvacuating(node)
}

func IsNodeRemovable(node *k8sCore.Node) bool {
	_, ok := poolUtil.FindLabel(node.Labels, commonNode.LabelKeyRemovable)
	return ok
}

func IsNodeUnremovable(node *k8sCore.Node) bool {
	_, ok := poolUtil.FindLabel(node.Labels, commonNode.LabelKeyUnremovable)
	return ok
}

// TODO There is no obvious way to determine if a node object corresponds to an existing node instance.
// We trust here that node GC or node graceful shutdown deal with it quickly enough.
func IsNodeTerminated(node *k8sCore.Node) bool {
	return false
}

func FindNodeResourcePool(node *k8sCore.Node) (string, bool) {
	return poolUtil.FindLabel(node.Labels, commonNode.LabelKeyResourcePool)
}

func FindNodeInstanceType(node *k8sCore.Node) (string, bool) {
	return poolUtil.FindLabel(node.Labels, commonNode.LabelKeyInstanceType)
}

func NodeBelongsToResourcePool(node *k8sCore.Node, resourcePool string) bool {
	return poolUtil.HasLabelAndValue(node.Labels, commonNode.LabelKeyResourcePool, resourcePool)
}

func Age(node *k8sCore.Node, now time.Time) time.Duration {
	return now.Sub(node.CreationTimestamp.Time)
}

func FromNodeToComputeResource(node *k8sCore.Node) poolApi.ComputeResource {
	return poolUtil.FromResourceListToComputeResource(node.Status.Allocatable)
}

func FromNodeToPhysicalComputeResource(node *k8sCore.Node, machinesByName map[string]*machineTypeV1.MachineTypeConfig) (*poolApi.ComputeResource, bool) {
	iType, ok := FindNodeInstanceType(node)
	if !ok {
		return nil, false
	}
	machineType, ok := machinesByName[iType]
	if !ok {
		return nil, false
	}

	return &poolApi.ComputeResource{
		CPU:         machineType.Spec.CPU,
		MemoryMB:    machineType.Spec.MemoryMB,
		DiskMB:      machineType.Spec.DiskMB,
		NetworkMBPS: machineType.Spec.NetworkMBPS,
		GPU:         machineType.Spec.GPU,
	}, true
}

func Names(nodes []*k8sCore.Node) []string {
	var names []string
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}

func FindTaint(node *k8sCore.Node, taintKey string) *k8sCore.Taint {
	for _, taint := range node.Spec.Taints {
		if taint.Key == taintKey {
			return &taint
		}
	}
	return nil
}

// Sort in place an array of nodes by the creation timestamp.
func SortNodesByAge(nodes []*k8sCore.Node) []*k8sCore.Node {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].CreationTimestamp.Before(&nodes[j].CreationTimestamp)
	})
	return nodes
}

func SumNodeResources(nodes []*k8sCore.Node) poolApi.ComputeResource {
	var sum poolApi.ComputeResource
	for _, node := range nodes {
		sum = sum.Add(FromNodeToComputeResource(node))
	}
	return sum
}

func SumNodeResourcesInMap(nodes map[string]*k8sCore.Node) poolApi.ComputeResource {
	var sum poolApi.ComputeResource
	for _, node := range nodes {
		sum = sum.Add(FromNodeToComputeResource(node))
	}
	return sum
}

func SumNodesPhysicalResources(nodes []*k8sCore.Node, machinesByName map[string]*machineTypeV1.MachineTypeConfig) (*poolApi.ComputeResource, bool) {
	var sum poolApi.ComputeResource
	for _, n := range nodes {
		r, ok := FromNodeToPhysicalComputeResource(n, machinesByName)
		if !ok {
			return nil, false
		}
		sum = sum.Add(*r)
	}
	return &sum, true
}
