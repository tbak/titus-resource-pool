package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
	"github.com/Netflix/titus-resource-pool/resourcepool"
	v1 "k8s.io/api/core/v1"
)

// `Usage` represents a resource consumption within a capacity group. The pod resource demand is adjusted to the
// capacity group resource shape. This is the worst case analysis.
// For example, if capacity group shape is {cpu=10, memory=20}, and one pod has allocation {cpu=5, memory=5}, and
// another one {cpu=5, memory=15}, it is rounded up respectively to {cpu=5, memory=10} and {cpu=8, memory=15}.
// We do this to avoid situations like this: node={cpu=100, memory=500}, pod1={cpu=1, memory=499}, pod2={cpu=99, memory=1}.
// If pod1 and pod2 are placed together on the same node they fit perfectly. But if not, they take two full nodes
// with little chance that another user can place something there. This example is an artificially constructed extreme
// case where real usage is a double of reservation, but in practice we can still expect this to be significant.
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
		}
	}

	resourcePool := snapshot.ResourcePool.Spec
	resourcePoolShape := resourcePool.ResourceShape.ComputeResource

	bufferShape := poolV1.Zero
	bufferTotal := poolV1.Zero
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
				inCapacityGroup[reservation.Spec.CapacityGroupName] = usage

				allReserved.Allocated = allReserved.Allocated.Add(usage.Allocated)
				allReserved.Unallocated = allReserved.Unallocated.Add(usage.Unallocated)
				allReserved.OverAllocation = allReserved.OverAllocation.Add(usage.OverAllocation)

				bufferAllocated, bufferOverallocation, elasticAllocated := buildBufferAndElasticUsage(remainingBuffer, bufferShape, resourcePoolShape, overallocatedPods)
				bufferAllocatedByCapacityGroup[reservation.Spec.CapacityGroupName] = bufferAllocated
				elasticAllocatedByCapacityGroup[reservation.Spec.CapacityGroupName] = elasticAllocated
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
	reservationShape := reservation.Spec.ComputeResource
	reservedResources := CapacityGroupResources(reservation)
	allocated := poolV1.ComputeResource{}
	overAllocated := poolV1.ComputeResource{}
	overAllocationPods := []*v1.Pod{}
	for _, pod := range snapshot.PodSnapshot.ScheduledByName {
		if poolPod.IsPodInCapacityGroup(pod, reservation.Name) {
			notAligned := poolPod.FromPodToComputeResource(pod)
			podResources := notAligned.AlignResourceRatios(reservationShape)
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

func buildBufferAndElasticUsage(remainingBuffer poolV1.ComputeResource, bufferShape poolV1.ComputeResource,
	resourcePoolShape poolV1.ComputeResource, bufferPods []*v1.Pod) (poolV1.ComputeResource, poolV1.ComputeResource, poolV1.ComputeResource) {
	bufferAllocated := poolV1.ComputeResource{}
	bufferOverallocation := poolV1.ComputeResource{}
	elasticAllocated := poolV1.ComputeResource{}
	for _, pod := range bufferPods {
		notAligned := poolPod.FromPodToComputeResource(pod)
		alignedToBuffer := notAligned.AlignResourceRatios(bufferShape)
		nextBufferAllocated := bufferAllocated.Add(alignedToBuffer)
		if nextBufferAllocated != remainingBuffer && !nextBufferAllocated.LessThan(remainingBuffer) {
			bufferOverallocation = bufferOverallocation.Add(alignedToBuffer)
			elasticAllocated = elasticAllocated.Add(notAligned.AlignResourceRatios(resourcePoolShape))
		} else {
			bufferAllocated = nextBufferAllocated
		}
	}
	return bufferAllocated, bufferOverallocation, elasticAllocated
}
