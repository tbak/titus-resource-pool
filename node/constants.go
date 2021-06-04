package node

const (
	// Labels
	NodeLabelBackend = "node.titus.netflix.com/backend"

	// Backends
	NodeBackendKubelet = "kubelet"

	// Possible node states
	NodeStateBootstrapping  = "bootstrapping"
	NodeStateActive         = "active"
	NodeStatePhasedOut      = "phasedOut"
	NodeStateDecommissioned = "decommissioned"
	NodeStateScalingDown    = "scalingDown"
	NodeStateBroken         = "broken"
	NodeStateRemovable      = "removable"
)

var NodeStatesAll = []string{
	NodeStateBootstrapping,
	NodeStateActive,
	NodeStatePhasedOut,
	NodeStateDecommissioned,
	NodeStateScalingDown,
	NodeStateBroken,
	NodeStateRemovable,
}
