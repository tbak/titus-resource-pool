package resourcepool

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	k8sCore "k8s.io/api/core/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

// Data structure that holds resource pool CRD and nodes and pods associated with this resource pool.
type ResourceSnapshot struct {
	// User provided
	client                 ctrlClient.Client
	ResourcePoolName       string
	NodeBootstrapThreshold time.Duration
	// State
	ResourcePool *poolV1.ResourcePoolConfig
	Machines     []*poolV1.MachineTypeConfig
	Nodes        []*k8sCore.Node
	Pods         []*k8sCore.Pod
	// Pods with the primary resource pool being this one
	PrimaryPods []*k8sCore.Pod
	// Helper data structures
	nodesByID map[string]*k8sCore.Node
}

func NewResourceSnapshot(client ctrlClient.Client, resourcePoolName string,
	nodeBootstrapThreshold time.Duration, withPods bool) (*ResourceSnapshot, error) {
	snapshot := ResourceSnapshot{
		client:                 client,
		ResourcePoolName:       resourcePoolName,
		NodeBootstrapThreshold: nodeBootstrapThreshold,
	}

	var err error
	if err = snapshot.ReloadResourcePool(); err != nil {
		return nil, err
	}
	if err = snapshot.ReloadMachines(); err != nil {
		return nil, err
	}
	if err = snapshot.ReloadNodes(); err != nil {
		return nil, err
	}
	if withPods {
		if err = snapshot.ReloadPods(); err != nil {
			return nil, err
		}
	} else {
		snapshot.Pods = []*k8sCore.Pod{}
	}
	return &snapshot, nil
}

// New resource snapshot that is statically configured. Reloading functions when called do nothing.
func NewStaticResourceSnapshot(resourcePool *poolV1.ResourcePoolConfig, machines []*poolV1.MachineTypeConfig,
	nodes []*k8sCore.Node, pods []*k8sCore.Pod, nodeBootstrapThreshold time.Duration) *ResourceSnapshot {
	snapshot := ResourceSnapshot{
		ResourcePoolName:       resourcePool.Name,
		ResourcePool:           resourcePool,
		NodeBootstrapThreshold: nodeBootstrapThreshold,
		Machines:               machines,
	}
	snapshot.updateNodeData(nodes)
	snapshot.updatePodData(pods)
	return &snapshot
}

func (snapshot *ResourceSnapshot) ActiveCapacity() poolV1.ComputeResource {
	total := poolV1.ComputeResource{}
	for _, node := range snapshot.Nodes {
		if !poolUtil.IsNodeOnItsWayOut(node) {
			total = total.Add(poolUtil.FromNodeToComputeResource(node))
		}
	}
	return total
}

func (snapshot *ResourceSnapshot) ActiveNodeCount() int64 {
	count := 0
	for _, node := range snapshot.Nodes {
		if !poolUtil.IsNodeOnItsWayOut(node) {
			count = count + 1
		}
	}
	return int64(count)
}

// Sum of resources of all nodes that are explicitly marked as decommissioned/removable.
func (snapshot *ResourceSnapshot) OnWayOutCapacity() poolV1.ComputeResource {
	total := poolV1.ComputeResource{}
	for _, node := range snapshot.Nodes {
		if poolUtil.IsNodeOnItsWayOut(node) {
			total = total.Add(poolUtil.FromNodeToComputeResource(node))
		}
	}
	return total
}

func (snapshot *ResourceSnapshot) OnWayOutNodeCount() int64 {
	count := 0
	for _, node := range snapshot.Nodes {
		if poolUtil.IsNodeOnItsWayOut(node) {
			count = count + 1
		}
	}
	return int64(count)
}

func (snapshot *ResourceSnapshot) NotProvisionedCapacity() poolV1.ComputeResource {
	return snapshot.ResourcePool.Spec.ResourceShape.Multiply(snapshot.ResourcePool.Spec.ResourceCount).
		SubWithLimit(snapshot.ActiveCapacity(), 0)
}

func (snapshot *ResourceSnapshot) NotProvisionedCount() int64 {
	return snapshot.NotProvisionedCapacity().SplitByWithCeil(snapshot.ResourcePool.Spec.ResourceShape.ComputeResource)
}

func (snapshot *ResourceSnapshot) FormatResourceSnapshot(options poolUtil.FormatterOptions) string {
	if options.Level == poolUtil.FormatCompact {
		return formatResourceSnapshotCompact(snapshot)
	} else if options.Level == poolUtil.FormatEssentials {
		return formatResourceSnapshotEssentials(snapshot)
	} else if options.Level == poolUtil.FormatDetails {
		return formatResourceSnapshotEssentials(snapshot)
	}
	return formatResourceSnapshotCompact(snapshot)
}

func (snapshot *ResourceSnapshot) DumpSnapshotToLog(log logr.Logger, options poolUtil.FormatterOptions,
	withNodes bool, withPods bool) {
	log.Info(fmt.Sprintf("Resource pool aggregates: %s", snapshot.FormatResourceSnapshot(options)))
	log.Info(fmt.Sprintf("Resource pool: %s", poolUtil.FormatResourcePool(snapshot.ResourcePool, options)))
	if withNodes {
		for _, node := range snapshot.Nodes {
			log.Info(fmt.Sprintf("Node: %s", poolUtil.FormatNode(node, snapshot.NodeBootstrapThreshold, options)))
		}
	}
	if withPods {
		for _, pod := range snapshot.Pods {
			log.Info(fmt.Sprintf("Pod: %s", poolUtil.FormatPod(pod, options)))
		}
	}
}

func (snapshot *ResourceSnapshot) AdjustResourcePoolSize(resourceCount int64) error {
	if snapshot.client == nil {
		snapshot.ResourcePool.Spec.ResourceCount = resourceCount
		return nil
	}

	update := snapshot.ResourcePool.DeepCopy()
	patch := ctrlClient.MergeFrom(update.DeepCopy())
	update.Spec.ResourceCount = resourceCount
	update.Spec.RequestedAt = time.Now().Unix()
	if err := snapshot.client.Patch(context.TODO(), update, patch); err != nil {
		return err
	}
	snapshot.ResourcePool = update
	return nil
}

func (snapshot *ResourceSnapshot) UpdateNode(nodeID string, transformer func(*k8sCore.Node)) error {
	node, ok := snapshot.nodesByID[nodeID]
	if !ok {
		return fmt.Errorf("resource pool does not include node %s", nodeID)
	}

	if snapshot.client == nil {
		transformer(node)
		return nil
	}

	patch := ctrlClient.MergeFrom(node.DeepCopy())
	transformer(node)
	if err := snapshot.client.Patch(context.TODO(), node, patch); err != nil {
		return err
	}
	return nil
}

func (snapshot *ResourceSnapshot) ReloadResourcePool() error {
	if snapshot.client == nil {
		return nil
	}

	resourcePool := poolV1.ResourcePoolConfig{}
	err := snapshot.client.Get(context.TODO(),
		ctrlClient.ObjectKey{Namespace: "default", Name: snapshot.ResourcePoolName}, &resourcePool)
	if err != nil {
		return fmt.Errorf("cannot read resource pool CRD: %s", snapshot.ResourcePoolName)
	}
	snapshot.ResourcePool = &resourcePool
	return nil
}

func (snapshot *ResourceSnapshot) ReloadMachines() error {
	if snapshot.client == nil {
		return nil
	}

	machineList := poolV1.MachineTypeConfigList{}
	if err := snapshot.client.List(context.TODO(), &machineList); err != nil {
		return errors.New("cannot read machine types")
	}

	var machines []*poolV1.MachineTypeConfig
	for _, machine := range machineList.Items {
		tmp := machine
		machines = append(machines, &tmp)
	}
	snapshot.Machines = machines
	return nil
}

func (snapshot *ResourceSnapshot) ReloadNodes() error {
	if snapshot.client == nil {
		return nil
	}

	nodeList := k8sCore.NodeList{}
	if err := snapshot.client.List(context.TODO(), &nodeList); err != nil {
		return errors.New("cannot read nodes")
	}
	snapshot.updateNodeData(poolNode.AsNodeReferenceList(&nodeList))
	return nil
}

func (snapshot *ResourceSnapshot) updateNodeData(current []*k8sCore.Node) {
	var nodes []*k8sCore.Node
	nodesByID := map[string]*k8sCore.Node{}
	for _, node := range current {
		if poolUtil.NodeBelongsToResourcePool(node, &snapshot.ResourcePool.Spec) {
			nodes = append(nodes, node)
			nodesByID[node.Name] = node
		}
	}

	snapshot.Nodes = nodes
	snapshot.nodesByID = nodesByID
}

func (snapshot *ResourceSnapshot) ReloadPods() error {
	if snapshot.client == nil {
		return nil
	}

	podList := k8sCore.PodList{}
	if err := snapshot.client.List(context.TODO(), &podList); err != nil {
		return errors.New("cannot read podList")
	}
	snapshot.updatePodData(poolPod.AsPodReferenceList(&podList))
	return nil
}

func (snapshot *ResourceSnapshot) updatePodData(current []*k8sCore.Pod) {
	var pods []*k8sCore.Pod
	for _, pod := range current {
		if poolUtil.PodBelongsToResourcePool(pod, &snapshot.ResourcePool.Spec, snapshot.Nodes) {
			pods = append(pods, pod)
		}
	}

	snapshot.Pods = pods
	snapshot.PrimaryPods = poolUtil.FindPodsWithPrimaryResourcePool(snapshot.ResourcePoolName, pods)
}

func formatResourceSnapshotCompact(snapshot *ResourceSnapshot) string {
	type Compact struct {
		Name                    string
		ActiveNodeCount         int64
		NotProvisionedNodeCount int64
		OnWayOutNodeCount       int64
	}
	value := Compact{
		Name:                    snapshot.ResourcePool.Name,
		ActiveNodeCount:         snapshot.ActiveNodeCount(),
		NotProvisionedNodeCount: snapshot.NotProvisionedCount(),
		OnWayOutNodeCount:       snapshot.OnWayOutNodeCount(),
	}
	return poolUtil.ToJSONString(value)
}

func formatResourceSnapshotEssentials(snapshot *ResourceSnapshot) string {
	type Compact struct {
		Name                    string
		ActiveNodeCount         int64
		NotProvisionedNodeCount int64
		OnWayOutNodeCount       int64
		ActiveResources         poolV1.ComputeResource
		NotProvisionedResources poolV1.ComputeResource
		OnWayOutResources       poolV1.ComputeResource
	}
	value := Compact{
		Name:                    snapshot.ResourcePool.Name,
		ActiveNodeCount:         snapshot.ActiveNodeCount(),
		NotProvisionedNodeCount: snapshot.NotProvisionedCount(),
		OnWayOutNodeCount:       snapshot.OnWayOutNodeCount(),
		ActiveResources:         snapshot.ActiveCapacity(),
		NotProvisionedResources: snapshot.NotProvisionedCapacity(),
		OnWayOutResources:       snapshot.OnWayOutCapacity(),
	}
	return poolUtil.ToJSONString(value)
}
