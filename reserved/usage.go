package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
	"github.com/Netflix/titus-resource-pool/resourcepool"
	v1 "k8s.io/api/core/v1"
)

// Usage represents a resource consumption within a capacity group.
type Usage struct {
	Allocated      poolV1.ComputeResource
	Unallocated    poolV1.ComputeResource
	OverAllocation poolV1.ComputeResource
}

type CapacityReservationUsage struct {
	// Reservation usage per capacity group. Buffer capacity group is not included here.
	InCapacityGroup map[string]Usage
	// Buffer capacity group name
	Buffer                         Usage
	BufferAllocatedByCapacityGroup map[string]poolV1.ComputeResource
	// Elastic
	Elastic                         Usage
	ElasticAllocatedByCapacityGroup map[string]poolV1.ComputeResource
	// Reservation usage for all capacity groups aggregated. Buffer usage is computed by taking over allocations from
	// all capacity groups. Buffer's `OverAllocation` is set to resources that could not be fit into the buffer.
	// Allocated and Unallocated is a sum of InCapacityGroup and Buffer.
	AllReserved Usage
}

func (u Usage) Add(other Usage) Usage {
	return Usage{
		Allocated:      u.Allocated.Add(other.Allocated),
		Unallocated:    u.Unallocated.Add(other.Unallocated),
		OverAllocation: u.OverAllocation.Add(other.OverAllocation),
	}
}

// For a given resource pool and reservations compute resource utilization per reservation.
// Only capacity groups associated with the given resource pool are considered.
func NewCapacityReservationUsage(snapshot *resourcepool.ResourceSnapshot,
	reservations []*capacityGroupV1.CapacityGroup, bufferName string) *CapacityReservationUsage {
	inCapacityGroup := map[string]Usage{}
	bufferAllocatedByCapacityGroup := map[string]poolV1.ComputeResource{}
	elasticAllocatedByCapacityGroup := map[string]poolV1.ComputeResource{}
	allReserved := Usage{}

	var bufferCapacityGroup *capacityGroupV1.CapacityGroup
	for _, reservation := range reservations {
		if reservation.Name == bufferName {
			bufferCapacityGroup = reservation
			break
		}
	}

	resourcePool := snapshot.ResourcePool.Spec

	var bufferShape poolV1.ComputeResource
	var bufferTotal poolV1.ComputeResource
	if bufferCapacityGroup != nil {
		bufferShape = bufferCapacityGroup.Spec.ComputeResource
		bufferTotal = bufferShape.Multiply(int64(bufferCapacityGroup.Spec.InstanceCount))
	}
	remainingBuffer := bufferTotal

	totalBufferOverallocation := poolV1.Zero
	totalElasticAllocation := poolV1.Zero
	for _, reservation := range reservations {
		if reservation.Spec.ResourcePoolName == snapshot.ResourcePoolName {
			if reservation.Name != bufferName {
				usage, overallocatedPods := buildUsage(snapshot, reservation)
				reservationName := reservation.Spec.OriginalName
				inCapacityGroup[reservationName] = usage

				allReserved.Allocated = allReserved.Allocated.Add(usage.Allocated)
				allReserved.Unallocated = allReserved.Unallocated.Add(usage.Unallocated)
				allReserved.OverAllocation = allReserved.OverAllocation.Add(usage.OverAllocation)

				bufferAllocated, bufferOverallocation, elasticAllocated := buildBufferAndElasticUsage(remainingBuffer, overallocatedPods)
				bufferAllocatedByCapacityGroup[reservationName] = bufferAllocated
				elasticAllocatedByCapacityGroup[reservationName] = elasticAllocated
				remainingBuffer = remainingBuffer.Sub(bufferAllocated)
				totalBufferOverallocation = totalBufferOverallocation.Add(bufferOverallocation)
				totalElasticAllocation = totalElasticAllocation.Add(elasticAllocated)
			}
		}
	}

	var bufferUsage Usage
	if bufferCapacityGroup != nil {
		bufferUsage.Allocated = bufferTotal.SubWithLimit(remainingBuffer, 0)
		bufferUsage.Unallocated = remainingBuffer
		bufferUsage.OverAllocation = totalBufferOverallocation

		allReserved.Allocated = allReserved.Allocated.Add(bufferUsage.Allocated)
		allReserved.Unallocated = allReserved.Unallocated.Add(bufferUsage.Unallocated)
		allReserved.OverAllocation = totalBufferOverallocation
	}

	totalElastic := resourcePool.ResourceShape.Multiply(resourcePool.ResourceCount).
		SubWithLimit(allReserved.Allocated.Add(allReserved.Unallocated), 0)
	elasticUsage := Usage{
		Allocated:   totalElasticAllocation,
		Unallocated: totalElastic.SubWithLimit(totalElasticAllocation, 0),
	}

	return &CapacityReservationUsage{
		InCapacityGroup:                 inCapacityGroup,
		Buffer:                          bufferUsage,
		BufferAllocatedByCapacityGroup:  bufferAllocatedByCapacityGroup,
		Elastic:                         elasticUsage,
		ElasticAllocatedByCapacityGroup: elasticAllocatedByCapacityGroup,
		AllReserved:                     allReserved,
	}
}

func buildUsage(snapshot *resourcepool.ResourceSnapshot, reservation *capacityGroupV1.CapacityGroup) (Usage, []*v1.Pod) {
	reservedResources := CapacityGroupResources(reservation)
	allocated := poolV1.ComputeResource{}
	overAllocated := poolV1.ComputeResource{}
	overAllocationPods := []*v1.Pod{}
	for _, pod := range snapshot.PodSnapshot.ScheduledByName {
		if poolPod.IsPodInCapacityGroup(pod, reservation) {
			podResources := poolPod.FromPodToComputeResource(pod)
			nextAllocated := allocated.Add(podResources)
			if nextAllocated != reservedResources && !nextAllocated.LessThan(reservedResources) {
				overAllocated = overAllocated.Add(podResources)
				overAllocationPods = append(overAllocationPods, pod)
			} else {
				allocated = nextAllocated
			}
		}
	}

	return Usage{
		Allocated:      allocated,
		Unallocated:    reservedResources.SubWithLimit(allocated, 0),
		OverAllocation: overAllocated,
	}, overAllocationPods
}

func buildBufferAndElasticUsage(remainingBuffer poolV1.ComputeResource,
	bufferPods []*v1.Pod) (poolV1.ComputeResource, poolV1.ComputeResource, poolV1.ComputeResource) {
	bufferAllocated := poolV1.ComputeResource{}
	bufferOverallocation := poolV1.ComputeResource{}
	elasticAllocated := poolV1.ComputeResource{}
	for _, pod := range bufferPods {
		podResources := poolPod.FromPodToComputeResource(pod)
		nextBufferAllocated := bufferAllocated.Add(podResources)
		if nextBufferAllocated != remainingBuffer && !nextBufferAllocated.LessThan(remainingBuffer) {
			bufferOverallocation = bufferOverallocation.Add(podResources)
			elasticAllocated = elasticAllocated.Add(podResources)
		} else {
			bufferAllocated = nextBufferAllocated
		}
	}
	return bufferAllocated, bufferOverallocation, elasticAllocated
}
