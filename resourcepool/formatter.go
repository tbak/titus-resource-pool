package resourcepool

import (
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func FormatResourcePool(resourcePool *poolV1.ResourcePoolConfig, options poolUtil.FormatterOptions) string {
	if options.Level == poolUtil.FormatCompact {
		return formatResourcePoolCompact(resourcePool)
	} else if options.Level == poolUtil.FormatEssentials {
		return formatResourcePoolEssentials(resourcePool)
	} else if options.Level == poolUtil.FormatDetails {
		return poolUtil.ToJSONString(resourcePool)
	}
	return formatResourcePoolCompact(resourcePool)
}

func formatResourcePoolCompact(pool *poolV1.ResourcePoolConfig) string {
	type Compact struct {
		Name               string
		ResourceCount      int64
		AutoScalingEnabled bool
	}
	value := Compact{
		Name:               pool.Name,
		ResourceCount:      pool.Spec.ResourceCount,
		AutoScalingEnabled: pool.Spec.ScalingRules.AutoScalingEnabled,
	}
	return poolUtil.ToJSONString(value)
}

func formatResourcePoolEssentials(pool *poolV1.ResourcePoolConfig) string {
	type Essentials struct {
		Name               string
		ResourceCount      int64
		ResourceShape      poolV1.ComputeResource
		AutoScalingEnabled bool
	}
	value := Essentials{
		Name:               pool.Name,
		ResourceCount:      pool.Spec.ResourceCount,
		ResourceShape:      pool.Spec.ResourceShape.ComputeResource,
		AutoScalingEnabled: pool.Spec.ScalingRules.AutoScalingEnabled,
	}
	return poolUtil.ToJSONString(value)
}
