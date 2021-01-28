package node

const (
	// Labels
	NodeLabelBackend = "node.titus.netflix.com/backend"

	// Backends
	NodeBackendKublet = "kubelet"

	// Possible node states
	NodeStateBootstrapping  = "bootstrapping"
	NodeStateActive         = "active"
	NodeStateDecommissioned = "decommissioned"
	NodeStateScalingDown    = "scalingDown"
	NodeStateBroken         = "broken"
	NodeStateRemovable      = "removable"
)

var NodeStatesAll = []string{
	NodeStateBootstrapping,
	NodeStateActive,
	NodeStateDecommissioned,
	NodeStateScalingDown,
	NodeStateBroken,
	NodeStateRemovable,
}
