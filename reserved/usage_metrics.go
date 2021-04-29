package reserved

import (
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	reserved = "reserved"
	buffer   = "buffer"
	elastic  = "elastic"
)

type UsageMetrics struct {
	resourcePoolName                       string
	bufferName                             string
	capacityGroupUsageUnrestricted         *metrics.GaugeVec
	capacityGroupUsageWithBufferAndElastic *metrics.GaugeVec
	totalReservedAndElasticUsage           *metrics.GaugeVec
	recentlyUpdatedCapacityGroups          map[string]string
}

func NewUsageMetrics(metricsSubsystem string, resourcePoolName string, bufferName string) *UsageMetrics {
	capacityGroupUsageUnrestricted := metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      metricsSubsystem,
			Name:           "capacityGroupUsageUnrestricted",
			Help:           "Capacity group resource usage with accounting the above reservation usage in the capacity group (%)",
			StabilityLevel: metrics.ALPHA,
		}, []string{"resourcePool", "capacityGroup", "used"},
	)
	capacityGroupUsageWithBufferAndElastic := metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      metricsSubsystem,
			Name:           "capacityGroupUsageWithBufferAndElastic",
			Help:           "Capacity group resource usage with excessive capacity group usage attributed to the buffer or elastic (%)",
			StabilityLevel: metrics.ALPHA,
		}, []string{"resourcePool", "capacityGroup", "resourceType"},
	)
	totalReservedAndElasticUsage := metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      metricsSubsystem,
			Name:           "totalReservedAndElasticUsage",
			Help:           "Total usage of resources split by reserved and elastic capacity (%)",
			StabilityLevel: metrics.ALPHA,
		}, []string{"resourcePool", "resourceType", "buffer", "used"},
	)

	legacyregistry.MustRegister(
		capacityGroupUsageUnrestricted,
		capacityGroupUsageWithBufferAndElastic,
		totalReservedAndElasticUsage,
	)

	return &UsageMetrics{
		resourcePoolName:                       resourcePoolName,
		bufferName:                             bufferName,
		capacityGroupUsageUnrestricted:         capacityGroupUsageUnrestricted,
		capacityGroupUsageWithBufferAndElastic: capacityGroupUsageWithBufferAndElastic,
		totalReservedAndElasticUsage:           totalReservedAndElasticUsage,
		recentlyUpdatedCapacityGroups:          map[string]string{},
	}
}

func (m *UsageMetrics) Update(usage *CapacityReservationUsage) {
	totalReserved := usage.AllReserved.Allocated.Add(usage.AllReserved.Unallocated)
	totalBuffer := usage.Buffer.Allocated.Add(usage.Buffer.Unallocated)
	totalElastic := usage.Elastic.Allocated.Add(usage.Elastic.Unallocated)

	updatedCapacityGroups := map[string]string{}
	for capacityGroupName, capacityGroupUsage := range usage.InCapacityGroup {
		total := capacityGroupUsage.Allocated.Add(capacityGroupUsage.Unallocated)
		unallocatedPercentage := capacityGroupUsage.Unallocated.MaxRatio(total) * 100

		// Here capacity group utilization can go above 100%
		unrestrictedAllocatedPercentage := capacityGroupUsage.Allocated.Add(capacityGroupUsage.OverAllocation).MaxRatio(total) * 100
		m.capacityGroupUsageUnrestricted.WithLabelValues(m.resourcePoolName, capacityGroupName, "true").Set(unrestrictedAllocatedPercentage)
		m.capacityGroupUsageUnrestricted.WithLabelValues(m.resourcePoolName, capacityGroupName, "false").Set(unallocatedPercentage)

		// Here excessive capacity group utilization > 100% is attribute to the buffer first, and elastic capacity second
		allocatedPercentage := capacityGroupUsage.Allocated.MaxRatio(total) * 100
		m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, capacityGroupName, reserved).Set(allocatedPercentage)

		if bufferUsage, ok := usage.BufferAllocatedByCapacityGroup[capacityGroupName]; ok {
			allocatedBufferPercentage := 0.0
			if totalBuffer.IsAnyAboveZero() {
				allocatedBufferPercentage = bufferUsage.MaxRatio(totalBuffer) * 100
			}
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, capacityGroupName, buffer).Set(allocatedBufferPercentage)
		} else {
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, capacityGroupName, buffer).Set(0)
		}

		if elasticUsage, ok := usage.ElasticAllocatedByCapacityGroup[capacityGroupName]; ok {
			allocatedElasticPercentage := 0.0
			if totalElastic.IsAnyAboveZero() {
				allocatedElasticPercentage = elasticUsage.MaxRatio(totalElastic) * 100
			}
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, capacityGroupName, elastic).Set(allocatedElasticPercentage)
		} else {
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, capacityGroupName, elastic).Set(0)
		}

		updatedCapacityGroups[capacityGroupName] = capacityGroupName
	}

	// Reset values for removed capacity groups
	for previousCapacityGroup := range m.recentlyUpdatedCapacityGroups {
		if _, ok := updatedCapacityGroups[previousCapacityGroup]; !ok {
			m.capacityGroupUsageUnrestricted.WithLabelValues(m.resourcePoolName, previousCapacityGroup, "true").Set(0)
			m.capacityGroupUsageUnrestricted.WithLabelValues(m.resourcePoolName, previousCapacityGroup, "false").Set(0)
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, previousCapacityGroup, reserved).Set(0)
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, previousCapacityGroup, buffer).Set(0)
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(m.resourcePoolName, previousCapacityGroup, elastic).Set(0)
		}
	}
	m.recentlyUpdatedCapacityGroups = updatedCapacityGroups

	// Total reserved and elastic.
	nonBufferAllocated := usage.AllReserved.Allocated.Sub(usage.Buffer.Allocated)
	nonBufferUnallocated := usage.AllReserved.Unallocated.Sub(usage.Buffer.Unallocated)
	m.totalReservedAndElasticUsage.WithLabelValues(m.resourcePoolName, reserved, "false", "true").Set(nonBufferAllocated.MaxRatio(totalReserved) * 100)
	m.totalReservedAndElasticUsage.WithLabelValues(m.resourcePoolName, reserved, "false", "false").Set(nonBufferUnallocated.MaxRatio(totalReserved) * 100)
	m.totalReservedAndElasticUsage.WithLabelValues(m.resourcePoolName, reserved, "true", "true").Set(usage.Buffer.Allocated.MaxRatio(totalReserved) * 100)
	m.totalReservedAndElasticUsage.WithLabelValues(m.resourcePoolName, reserved, "true", "false").Set(usage.Buffer.Unallocated.MaxRatio(totalReserved) * 100)

	elasticPercentage := usage.Elastic.Allocated.MaxRatio(totalElastic) * 100
	m.totalReservedAndElasticUsage.WithLabelValues(m.resourcePoolName, elastic, "false", "true").Set(elasticPercentage)
	m.totalReservedAndElasticUsage.WithLabelValues(m.resourcePoolName, elastic, "false", "false").Set(100 - elasticPercentage)
}
