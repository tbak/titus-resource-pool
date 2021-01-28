package node

const (
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
