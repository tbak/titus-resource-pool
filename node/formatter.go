package node

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	"time"

	v1 "k8s.io/api/core/v1"

	. "github.com/Netflix/titus-resource-pool/util"
)

func FormatNode(node *v1.Node, ageThreshold time.Duration, options FormatterOptions) string {
	if options.Level == FormatCompact {
		return formatNodeCompact(node, ageThreshold)
	} else if options.Level == FormatEssentials {
		return formatNodeEssentials(node, ageThreshold)
	} else if options.Level == FormatDetails {
		return ToJSONString(node)
	}
	return formatNodeCompact(node, ageThreshold)
}

func formatNodeCompact(node *v1.Node, ageThreshold time.Duration) string {
	type Compact struct {
		Name     string
		Up       bool
		OnWayOut bool
	}
	value := Compact{
		Name:     node.Name,
		Up:       IsNodeAvailableForScheduling(node, time.Now(), ageThreshold),
		OnWayOut: IsNodeOnItsWayOut(node),
	}
	return ToJSONString(value)
}

func formatNodeEssentials(node *v1.Node, ageThreshold time.Duration) string {
	type Compact struct {
		Name               string
		Up                 bool
		OnWayOut           bool
		AvailableResources poolV1.ComputeResource
	}
	value := Compact{
		Name:               node.Name,
		Up:                 IsNodeAvailableForScheduling(node, time.Now(), ageThreshold),
		OnWayOut:           IsNodeOnItsWayOut(node),
		AvailableResources: FromNodeToComputeResource(node),
	}
	return ToJSONString(value)
}
