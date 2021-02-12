package node

import (
	"time"

	v1 "k8s.io/api/core/v1"

	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"

	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func FormatNode(node *v1.Node, ageThreshold time.Duration, options poolUtil.FormatterOptions) string {
	if options.Level == poolUtil.FormatCompact {
		return formatNodeCompact(node, ageThreshold)
	} else if options.Level == poolUtil.FormatEssentials {
		return formatNodeEssentials(node, ageThreshold)
	} else if options.Level == poolUtil.FormatDetails {
		return poolUtil.ToJSONString(node)
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
	return poolUtil.ToJSONString(value)
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
	return poolUtil.ToJSONString(value)
}
