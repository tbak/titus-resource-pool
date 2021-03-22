package pod

import (
	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolNode "github.com/Netflix/titus-resource-pool/node"
	k8sCore "k8s.io/api/core/v1"
	"time"
)

type Metadata struct {
	PrimaryResourcePool   string
	AssignedResourcePools []string
	UsedResourcePool      string
	PodResources          v1.ComputeResource
}

// Node data snapshot with useful indexes for fast access. Snapshot struct can be mutated by calling the provided
// functions (Add, Transform). Those updates are applied in place, so if a client keeps reference to a collection
// (for example AllByName), it may change as well.
// To support fast O(1) mutations, only map collections are provided.
type Snapshot struct {
	AllByName         map[string]*k8sCore.Pod
	QueuedYoungByName map[string]*k8sCore.Pod
	QueuedOldByName   map[string]*k8sCore.Pod
	ScheduledByName   map[string]*k8sCore.Pod
	FinishedByName    map[string]*k8sCore.Pod
	Metadata          map[string]*Metadata
	// Pods with the primary resource pool being this one
	Primary map[string]*k8sCore.Pod
}

type Options struct {
	SupportGPUs        bool
	PastYoungThreshold func(pod *k8sCore.Pod, now time.Time) bool
}

func NewEmpty() *Snapshot {
	return &Snapshot{
		AllByName:         map[string]*k8sCore.Pod{},
		QueuedYoungByName: map[string]*k8sCore.Pod{},
		QueuedOldByName:   map[string]*k8sCore.Pod{},
		ScheduledByName:   map[string]*k8sCore.Pod{},
		FinishedByName:    map[string]*k8sCore.Pod{},
		Metadata:          map[string]*Metadata{},
		Primary:           map[string]*k8sCore.Pod{},
	}
}

// Returns Snapshot of pods associated with the given resource pool and the list of the remaining pods.
// This factory does not filter out running pods on node resources not belonging to this resource pool
// (expected if a pod is associated with many resource pools). This additional filtering step can be done by calling
// NewFilteredByNodeAllocation and passing the node data.
func NewSnapshotOfResourcePool(pods []*k8sCore.Pod, resourcePool string, options Options) (*Snapshot, []*k8sCore.Pod) {
	now := time.Now()
	result := NewEmpty()

	var pastYoungThreshold func(pod *k8sCore.Pod) bool
	if options.PastYoungThreshold == nil {
		pastYoungThreshold = func(pod *k8sCore.Pod) bool {
			return true
		}
	} else {
		pastYoungThreshold = func(pod *k8sCore.Pod) bool {
			return options.PastYoungThreshold(pod, now)
		}
	}

	other := []*k8sCore.Pod{}

	for _, pod := range pods {
		if metadata, ok := buildPodMetadata(pod, resourcePool, options); !ok {
			other = append(other, pod)
		} else {
			result.AllByName[pod.Name] = pod
			result.Metadata[pod.Name] = metadata
			if IsPodWaitingToBeScheduled(pod) {
				if pastYoungThreshold(pod) {
					result.QueuedOldByName[pod.Name] = pod
				} else {
					result.QueuedYoungByName[pod.Name] = pod
				}
			} else if IsPodRunning(pod) {
				result.ScheduledByName[pod.Name] = pod
			} else if IsPodFinished(pod) {
				result.FinishedByName[pod.Name] = pod
			}
			if metadata.PrimaryResourcePool == resourcePool {
				result.Primary[pod.Name] = pod
			}
		}
	}
	return result, other
}

// Given the unfilteredSnapshot, remove all pods in running state that run on nodes not owned by the  given resource pool.
func NewFilteredByNodeAllocation(unfilteredSnapshot *Snapshot, resourcePool string, nodeSnapshot *poolNode.Snapshot) (*Snapshot, []*k8sCore.Pod) {
	filtered := NewEmpty()
	other := []*k8sCore.Pod{}
	for _, pod := range unfilteredSnapshot.AllByName {
		if shouldFilterOutPod(pod, resourcePool, nodeSnapshot) {
			other = append(other, pod)
		} else {
			filtered.AllByName[pod.Name] = pod
			filtered.Metadata[pod.Name] = unfilteredSnapshot.Metadata[pod.Name]
			if _, ok := unfilteredSnapshot.QueuedYoungByName[pod.Name]; ok {
				filtered.QueuedYoungByName[pod.Name] = pod
			} else if _, ok := unfilteredSnapshot.QueuedOldByName[pod.Name]; ok {
				filtered.QueuedOldByName[pod.Name] = pod
			} else if _, ok := unfilteredSnapshot.ScheduledByName[pod.Name]; ok {
				filtered.ScheduledByName[pod.Name] = pod
			} else if _, ok := unfilteredSnapshot.FinishedByName[pod.Name]; ok {
				filtered.FinishedByName[pod.Name] = pod
			}

			if _, ok := unfilteredSnapshot.Primary[pod.Name]; ok {
				filtered.Primary[pod.Name] = pod
			}
		}
	}
	return filtered, other
}

func shouldFilterOutPod(pod *k8sCore.Pod, resourcePool string, nodeSnapshot *poolNode.Snapshot) bool {
	if pod.Spec.NodeName == "" {
		return false
	}
	// If the pod is assigned to a node, we check that the node itself belongs to the same resource pool.
	if nodeMetadata, ok := nodeSnapshot.MetadataByteName[pod.Spec.NodeName]; ok {
		if nodeMetadata.ResourcePool == resourcePool {
			return false
		}
	}
	return true
}

func (s *Snapshot) IsPodWaitingToBeScheduled(podName string) bool {
	if pod, ok := s.AllByName[podName]; ok {
		return IsPodWaitingToBeScheduled(pod)
	}
	return false
}

func buildPodMetadata(pod *k8sCore.Pod, resourcePool string, options Options) (*Metadata, bool) {
	podResourcePools, ok := FindPodAssignedResourcePools(pod)
	if !ok {
		return nil, false
	}

	podResources := FromPodToComputeResource(pod)

	// Do not look at pods requesting GPU resources, but running in non-GPU resource pool.
	if !options.SupportGPUs && podResources.GPU > 0 {
		return nil, false
	}

	metadata := Metadata{
		PrimaryResourcePool:   podResourcePools[0],
		AssignedResourcePools: podResourcePools,
		PodResources:          podResources,
	}

	for _, pool := range podResourcePools {
		if pool == resourcePool {
			return &metadata, true
		}
	}

	return nil, false
}
