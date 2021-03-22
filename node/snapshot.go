package node

import (
	"fmt"
	"time"

	k8sCore "k8s.io/api/core/v1"

	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

type Metadata struct {
	ResourcePool  string
	NodeResources v1.ComputeResource
}

// Node data snapshot with useful indexes for fast access. Snapshot struct can be mutated by calling the provided
// functions (Add, Transform). Those updates are applied in place, so if a client keeps reference to a collection
// (for example AllByName), it may change as well.
// To support fast O(1) mutations, only map collections are provided.
type Snapshot struct {
	AllByName           map[string]*k8sCore.Node
	BootstrappingByName map[string]*k8sCore.Node
	ActiveByName        map[string]*k8sCore.Node
	OnWayOutByName      map[string]*k8sCore.Node
	MetadataByteName    map[string]*Metadata
	// Explicitly excluded nodes which otherwise would be in AllByName and one of the BootstrappingByName, ActiveByName
	// or OnWayOutByName collections. Primary use case is to exclude nodes running experimental Kube backends.
	ExcludedByName map[string]*k8sCore.Node
	// Internal state
	options Options
}

type Options struct {
	// A predicate telling if a node is past its bootstrap stage.
	PastBootstrapDeadline func(node *k8sCore.Node, now time.Time) bool
	// A predicate for identifying nodes to be excluded.
	Exclude func(node *k8sCore.Node) bool
}

func NewEmptySnapshot() *Snapshot {
	return &Snapshot{
		AllByName:           map[string]*k8sCore.Node{},
		BootstrappingByName: map[string]*k8sCore.Node{},
		ActiveByName:        map[string]*k8sCore.Node{},
		OnWayOutByName:      map[string]*k8sCore.Node{},
		MetadataByteName:    map[string]*Metadata{},
		ExcludedByName:      map[string]*k8sCore.Node{},
		options:             Options{},
	}
}

// Returns Snapshot of nodes associated with the given resource pool and the list of the remaining nodes.
func NewSnapshotOfResourcePool(nodes []*k8sCore.Node, resourcePool string, options Options) (*Snapshot, []*k8sCore.Node) {
	now := time.Now()
	pastBootstrapDeadline := currentPastBootstrapDeadline(options, now)

	result := NewEmptySnapshot()
	result.options = options

	other := []*k8sCore.Node{}
	for _, node := range nodes {
		if poolName, ok := FindNodeResourcePool(node); ok && poolName == resourcePool {
			if options.Exclude != nil && options.Exclude(node) {
				result.ExcludedByName[node.Name] = node
			} else {
				result.AllByName[node.Name] = node
				result.MetadataByteName[node.Name] = buildMetadata(node)
				if IsNodeOnItsWayOut(node) {
					result.OnWayOutByName[node.Name] = node
				} else if IsNodeBootstrapping2(node, pastBootstrapDeadline) {
					result.BootstrappingByName[node.Name] = node
				} else {
					result.ActiveByName[node.Name] = node
				}
			}
		} else {
			other = append(other, node)
		}
	}
	return result, other
}

func buildMetadata(node *k8sCore.Node) *Metadata {
	resourcePool, _ := FindNodeResourcePool(node)
	return &Metadata{
		ResourcePool:  resourcePool,
		NodeResources: FromNodeToComputeResource(node),
	}
}

func currentPastBootstrapDeadline(options Options, now time.Time) func(node *k8sCore.Node) bool {
	var pastBootstrapDeadline func(node *k8sCore.Node) bool
	if options.PastBootstrapDeadline == nil {
		pastBootstrapDeadline = func(node *k8sCore.Node) bool {
			return true
		}
	} else {
		pastBootstrapDeadline = func(node *k8sCore.Node) bool {
			return options.PastBootstrapDeadline(node, now)
		}
	}
	return pastBootstrapDeadline
}

// Add a node. If a node already exists, it is overridden. Returns true if the node was not in the snapshot yet.
func (s *Snapshot) Add(node *k8sCore.Node) bool {
	_, found := s.AllByName[node.Name]
	if !found {
		_, found = s.ExcludedByName[node.Name]
	}

	if s.options.Exclude != nil && s.options.Exclude(node) {
		s.ExcludedByName[node.Name] = node
		delete(s.AllByName, node.Name)
		delete(s.BootstrappingByName, node.Name)
		delete(s.ActiveByName, node.Name)
		delete(s.OnWayOutByName, node.Name)
		delete(s.MetadataByteName, node.Name)
		return found
	}

	pastBootstrapDeadline := currentPastBootstrapDeadline(s.options, time.Now())

	delete(s.ExcludedByName, node.Name)
	delete(s.BootstrappingByName, node.Name)
	delete(s.ActiveByName, node.Name)
	delete(s.OnWayOutByName, node.Name)

	s.AllByName[node.Name] = node
	s.MetadataByteName[node.Name] = buildMetadata(node)
	if IsNodeOnItsWayOut(node) {
		s.OnWayOutByName[node.Name] = node
	} else if IsNodeBootstrapping2(node, pastBootstrapDeadline) {
		s.BootstrappingByName[node.Name] = node
	} else {
		s.ActiveByName[node.Name] = node
	}

	return found
}

func (s *Snapshot) Transform(nodeName string, transformer func(*k8sCore.Node)) (*k8sCore.Node, error) {
	node, ok := s.AllByName[nodeName]
	if !ok {
		return nil, fmt.Errorf("node snapshot does not include node %s", nodeName)
	}
	// We mutate the object itself, but we must add it again to make sure indexes are updated.
	transformer(node)
	s.Add(node)
	return node, nil
}

func (s *Snapshot) ContainsName(nodeName string) bool {
	if _, ok := s.AllByName[nodeName]; ok {
		return true
	}
	return false
}
