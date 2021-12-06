package reserved

import (
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	reserved = "reserved"
	buffer   = "buffer"
	elastic  = "elastic"
)

type usageMetricsInternal struct {
	capacityGroupUsageUnrestricted         *metrics.GaugeVec
	capacityGroupUsageWithBufferAndElastic *metrics.GaugeVec
	totalReservedAndElasticUsage           *metrics.GaugeVec
}

type UsageMetrics struct {
	resourcePoolName                       string
	bufferName                             string
	capacityGroupUsageUnrestricted         *prometheus.GaugeVec
	capacityGroupUsageWithBufferAndElastic *prometheus.GaugeVec
	totalReservedAndElasticUsage           *prometheus.GaugeVec
	recentlyUpdatedCapacityGroups          map[string]string
}

// Duplicate metrics registrations are not allowed by prometheus, so we have to track it globally.
var (
	usageMetricsRegistryLock sync.Mutex
	usageMetricsRegistry     = map[string]*usageMetricsInternal{}
)

func NewUsageMetrics(metricsSubsystem string, resourcePoolName string, bufferName string, leader bool) *UsageMetrics {
	internalMetrics := getOrCreateInternalMetrics(metricsSubsystem)
	sharedLabels := prometheus.Labels{
		"leader":       strconv.FormatBool(leader),
		"resourcePool": resourcePoolName,
	}

	m := &UsageMetrics{
		resourcePoolName:                       resourcePoolName,
		bufferName:                             bufferName,
		capacityGroupUsageUnrestricted:         internalMetrics.capacityGroupUsageUnrestricted.MustCurryWith(sharedLabels),
		capacityGroupUsageWithBufferAndElastic: internalMetrics.capacityGroupUsageWithBufferAndElastic.MustCurryWith(sharedLabels),
		totalReservedAndElasticUsage:           internalMetrics.totalReservedAndElasticUsage.MustCurryWith(sharedLabels),
		recentlyUpdatedCapacityGroups:          map[string]string{},
	}
	return m
}

func getOrCreateInternalMetrics(metricsSubsystem string) *usageMetricsInternal {
	usageMetricsRegistryLock.Lock()
	defer usageMetricsRegistryLock.Unlock()

	if subsystemMetrics, ok := usageMetricsRegistry[metricsSubsystem]; ok {
		return subsystemMetrics
	}

	capacityGroupUsageUnrestricted := metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      metricsSubsystem,
			Name:           "capacityGroupUsageUnrestricted",
			Help:           "Capacity group resource usage with accounting the above reservation usage in the capacity group (%)",
			StabilityLevel: metrics.ALPHA,
		}, []string{"leader", "resourcePool", "capacityGroup", "used"},
	)
	capacityGroupUsageWithBufferAndElastic := metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      metricsSubsystem,
			Name:           "capacityGroupUsageWithBufferAndElastic",
			Help:           "Capacity group resource usage with excessive capacity group usage attributed to the buffer or elastic (%)",
			StabilityLevel: metrics.ALPHA,
		}, []string{"leader", "resourcePool", "capacityGroup", "resourceType"},
	)
	totalReservedAndElasticUsage := metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      metricsSubsystem,
			Name:           "totalReservedAndElasticUsage",
			Help:           "Total usage of resources split by reserved and elastic capacity (%)",
			StabilityLevel: metrics.ALPHA,
		}, []string{"leader", "resourcePool", "resourceType", "buffer", "used"},
	)

	legacyregistry.MustRegister(
		capacityGroupUsageUnrestricted,
		capacityGroupUsageWithBufferAndElastic,
		totalReservedAndElasticUsage,
	)

	resourcePoolMetrics := &usageMetricsInternal{
		capacityGroupUsageUnrestricted:         capacityGroupUsageUnrestricted,
		capacityGroupUsageWithBufferAndElastic: capacityGroupUsageWithBufferAndElastic,
		totalReservedAndElasticUsage:           totalReservedAndElasticUsage,
	}
	usageMetricsRegistry[metricsSubsystem] = resourcePoolMetrics
	return resourcePoolMetrics
}

func (m *UsageMetrics) Reset() {
	m.capacityGroupUsageUnrestricted.Reset()
	m.capacityGroupUsageWithBufferAndElastic.Reset()
	m.totalReservedAndElasticUsage.Reset()
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
		m.capacityGroupUsageUnrestricted.WithLabelValues(capacityGroupName, "true").Set(unrestrictedAllocatedPercentage)
		m.capacityGroupUsageUnrestricted.WithLabelValues(capacityGroupName, "false").Set(unallocatedPercentage)

		// Here excessive capacity group utilization > 100% is attribute to the buffer first, and elastic capacity second
		allocatedPercentage := capacityGroupUsage.Allocated.MaxRatio(total) * 100
		m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(capacityGroupName, reserved).Set(allocatedPercentage)

		if bufferUsage, ok := usage.BufferAllocatedByCapacityGroup[capacityGroupName]; ok {
			allocatedBufferPercentage := 0.0
			if totalBuffer.IsAnyAboveZero() {
				allocatedBufferPercentage = bufferUsage.MaxRatio(totalBuffer) * 100
			}
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(capacityGroupName, buffer).Set(allocatedBufferPercentage)
		} else {
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(capacityGroupName, buffer).Set(0)
		}

		if elasticUsage, ok := usage.ElasticAllocatedByCapacityGroup[capacityGroupName]; ok {
			allocatedElasticPercentage := 0.0
			if totalElastic.IsAnyAboveZero() {
				allocatedElasticPercentage = elasticUsage.MaxRatio(totalElastic) * 100
			}
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(capacityGroupName, elastic).Set(allocatedElasticPercentage)
		} else {
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(capacityGroupName, elastic).Set(0)
		}

		updatedCapacityGroups[capacityGroupName] = capacityGroupName
	}

	// Reset values for removed capacity groups
	for previousCapacityGroup := range m.recentlyUpdatedCapacityGroups {
		if _, ok := updatedCapacityGroups[previousCapacityGroup]; !ok {
			m.capacityGroupUsageUnrestricted.WithLabelValues(previousCapacityGroup, "true").Set(0)
			m.capacityGroupUsageUnrestricted.WithLabelValues(previousCapacityGroup, "false").Set(0)
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(previousCapacityGroup, reserved).Set(0)
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(previousCapacityGroup, buffer).Set(0)
			m.capacityGroupUsageWithBufferAndElastic.WithLabelValues(previousCapacityGroup, elastic).Set(0)
		}
	}
	m.recentlyUpdatedCapacityGroups = updatedCapacityGroups

	// Total reserved and elastic.
	nonBufferAllocated := usage.AllReserved.Allocated.Sub(usage.Buffer.Allocated)
	nonBufferUnallocated := usage.AllReserved.Unallocated.Sub(usage.Buffer.Unallocated)
	m.totalReservedAndElasticUsage.WithLabelValues(reserved, "false", "true").Set(nonBufferAllocated.MaxRatio(totalReserved) * 100)
	m.totalReservedAndElasticUsage.WithLabelValues(reserved, "false", "false").Set(nonBufferUnallocated.MaxRatio(totalReserved) * 100)
	m.totalReservedAndElasticUsage.WithLabelValues(reserved, "true", "true").Set(usage.Buffer.Allocated.MaxRatio(totalReserved) * 100)
	m.totalReservedAndElasticUsage.WithLabelValues(reserved, "true", "false").Set(usage.Buffer.Unallocated.MaxRatio(totalReserved) * 100)

	elasticPercentage := usage.Elastic.Allocated.MaxRatio(totalElastic) * 100
	m.totalReservedAndElasticUsage.WithLabelValues(elastic, "false", "true").Set(elasticPercentage)
	m.totalReservedAndElasticUsage.WithLabelValues(elastic, "false", "false").Set(100 - elasticPercentage)
}
