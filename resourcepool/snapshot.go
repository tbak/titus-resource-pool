package resourcepool

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	k8sCore "k8s.io/api/core/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	machineTypeV1 "github.com/Netflix/titus-controllers-api/api/machinetype/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolMachine "github.com/Netflix/titus-resource-pool/machine"
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
	PodYoungThreshold      time.Duration
	IncludeKubeletBackend  bool
	// State
	ResourcePool   *poolV1.ResourcePoolConfig
	Machines       []*machineTypeV1.MachineTypeConfig
	MachinesByName map[string]*machineTypeV1.MachineTypeConfig
	NodeSnapshot   *poolNode.Snapshot
	PodSnapshot    *poolPod.Snapshot
}

func NewResourceSnapshot(client ctrlClient.Client, resourcePoolName string,
	nodeBootstrapThreshold time.Duration, includeKubeletBackend bool, withPods bool) (*ResourceSnapshot, error) {
	snapshot := ResourceSnapshot{
		client:                 client,
		ResourcePoolName:       resourcePoolName,
		NodeBootstrapThreshold: nodeBootstrapThreshold,
		IncludeKubeletBackend:  includeKubeletBackend,
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
		snapshot.PodSnapshot = poolPod.NewEmpty()
	}
	return &snapshot, nil
}

// New resource snapshot that is statically configured. Reloading functions when called do nothing.
func NewStaticResourceSnapshot(resourcePool *poolV1.ResourcePoolConfig, machines []*machineTypeV1.MachineTypeConfig,
	nodes []*k8sCore.Node, pods []*k8sCore.Pod, nodeBootstrapThreshold time.Duration, podYoungThreshold time.Duration,
	includeKubeletBackend bool) *ResourceSnapshot {
	snapshot := ResourceSnapshot{
		ResourcePoolName:       resourcePool.Name,
		ResourcePool:           resourcePool,
		NodeBootstrapThreshold: nodeBootstrapThreshold,
		PodYoungThreshold:      podYoungThreshold,
		IncludeKubeletBackend:  includeKubeletBackend,
		Machines:               machines,
		MachinesByName:         poolMachine.AsMachineTypeMap(machines),
	}
	snapshot.updateNodeData(nodes)
	snapshot.updatePodData(pods)
	return &snapshot
}

// New resource snapshot that is statically configured. Reloading functions when called do nothing.
func NewStaticResourceSnapshot2(resourcePool *poolV1.ResourcePoolConfig, machines []*machineTypeV1.MachineTypeConfig,
	nodeSnapshot *poolNode.Snapshot, podSnapshot *poolPod.Snapshot, nodeBootstrapThreshold time.Duration,
	podYoungThreshold time.Duration, includeKubeletBackend bool) *ResourceSnapshot {
	snapshot := ResourceSnapshot{
		ResourcePoolName:       resourcePool.Name,
		ResourcePool:           resourcePool,
		NodeBootstrapThreshold: nodeBootstrapThreshold,
		PodYoungThreshold:      podYoungThreshold,
		IncludeKubeletBackend:  includeKubeletBackend,
		Machines:               machines,
		MachinesByName:         poolMachine.AsMachineTypeMap(machines),
		NodeSnapshot:           nodeSnapshot,
		PodSnapshot:            podSnapshot,
	}
	return &snapshot
}

func (snapshot *ResourceSnapshot) ActiveCapacity() poolV1.ComputeResource {
	return poolNode.SumNodeResourcesInMap(snapshot.NodeSnapshot.ActiveByName)
}

func (snapshot *ResourceSnapshot) ActiveNodeCount() int64 {
	return int64(len(snapshot.NodeSnapshot.ActiveByName))
}

// Sum of resources of all nodes that are explicitly marked as decommissioned/removable.
func (snapshot *ResourceSnapshot) OnWayOutCapacity() poolV1.ComputeResource {
	return poolNode.SumNodeResourcesInMap(snapshot.NodeSnapshot.OnWayOutByName)
}

func (snapshot *ResourceSnapshot) OnWayOutNodeCount() int64 {
	return int64(len(snapshot.NodeSnapshot.OnWayOutByName))
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
	log.Info(fmt.Sprintf("Resource pool: %s", FormatResourcePool(snapshot.ResourcePool, options)))
	if withNodes {
		for _, node := range snapshot.NodeSnapshot.AllByName {
			log.Info(fmt.Sprintf("Node: %s", poolNode.FormatNode(node, snapshot.NodeBootstrapThreshold, options)))
		}
	}
	if withPods {
		for _, pod := range snapshot.PodSnapshot.AllByName {
			log.Info(fmt.Sprintf("Pod: %s", poolPod.FormatPod(pod, options)))
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
	node, ok := snapshot.NodeSnapshot.AllByName[nodeID]
	var patch ctrlClient.Patch
	if ok && snapshot.client != nil {
		patch = ctrlClient.MergeFrom(node.DeepCopy())
	}

	if _, err := snapshot.NodeSnapshot.Transform(nodeID, transformer); err != nil {
		return err
	}

	if snapshot.client != nil {
		if err := snapshot.client.Patch(context.TODO(), node, patch); err != nil {
			return err
		}
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

	machineList := machineTypeV1.MachineTypeConfigList{}
	if err := snapshot.client.List(context.TODO(), &machineList); err != nil {
		return errors.New("cannot read machine types")
	}

	var machines []*machineTypeV1.MachineTypeConfig
	for _, machine := range machineList.Items {
		tmp := machine
		machines = append(machines, &tmp)
	}
	snapshot.Machines = machines
	snapshot.MachinesByName = poolMachine.AsMachineTypeMap(machines)
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
	snapshot.NodeSnapshot, _ = poolNode.NewSnapshotOfResourcePool(current, snapshot.ResourcePoolName, snapshot.MachinesByName,
		poolNode.Options{
			PastBootstrapDeadline: func(node *k8sCore.Node, now time.Time) bool {
				return poolNode.Age(node, now) > snapshot.NodeBootstrapThreshold
			},
			Exclude: func(node *k8sCore.Node) bool {
				return !snapshot.IncludeKubeletBackend && poolNode.IsKubeletNode(node)
			},
		})
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
	unfiltered, _ := poolPod.NewSnapshotOfResourcePool(current, snapshot.ResourcePoolName, poolPod.Options{
		SupportGPUs: snapshot.ResourcePool.Spec.ResourceShape.GPU > 0,
		PastYoungThreshold: func(pod *k8sCore.Pod, now time.Time) bool {
			return poolPod.Age(pod, now) > snapshot.PodYoungThreshold
		},
	})
	snapshot.PodSnapshot, _ = poolPod.NewFilteredByNodeAllocation(unfiltered, snapshot.ResourcePoolName, snapshot.NodeSnapshot)
}

func formatResourceSnapshotCompact(snapshot *ResourceSnapshot) string {
	type Compact struct {
		Name                    string
		ActiveNodeCount         int64
		NotProvisionedNodeCount int64
		OnWayOutNodeCount       int64
		ExcludedNodeCount       int64
	}
	value := Compact{
		Name:                    snapshot.ResourcePool.Name,
		ActiveNodeCount:         snapshot.ActiveNodeCount(),
		NotProvisionedNodeCount: snapshot.NotProvisionedCount(),
		OnWayOutNodeCount:       snapshot.OnWayOutNodeCount(),
		ExcludedNodeCount:       int64(len(snapshot.NodeSnapshot.ExcludedByName)),
	}
	return poolUtil.ToJSONString(value)
}

func formatResourceSnapshotEssentials(snapshot *ResourceSnapshot) string {
	type Compact struct {
		Name                    string
		ActiveNodeCount         int64
		NotProvisionedNodeCount int64
		OnWayOutNodeCount       int64
		ExcludedNodeCount       int64
		ActiveResources         poolV1.ComputeResource
		NotProvisionedResources poolV1.ComputeResource
		OnWayOutResources       poolV1.ComputeResource
	}
	value := Compact{
		Name:                    snapshot.ResourcePool.Name,
		ActiveNodeCount:         snapshot.ActiveNodeCount(),
		NotProvisionedNodeCount: snapshot.NotProvisionedCount(),
		OnWayOutNodeCount:       snapshot.OnWayOutNodeCount(),
		ExcludedNodeCount:       int64(len(snapshot.NodeSnapshot.ExcludedByName)),
		ActiveResources:         snapshot.ActiveCapacity(),
		NotProvisionedResources: snapshot.NotProvisionedCapacity(),
		OnWayOutResources:       snapshot.OnWayOutCapacity(),
	}
	return poolUtil.ToJSONString(value)
}
