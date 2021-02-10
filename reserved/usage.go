package reserved

import (
	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
	"github.com/Netflix/titus-resource-pool/resourcepool"
)

type Usage struct {
	Allocated   v1.ComputeResource
	Unallocated v1.ComputeResource
}

type CapacityReservationUsage struct {
	// Reservation usage per capacity group.
	InCapacityGroup map[string]Usage
	// Reservation usage for all capacity groups aggregated.
	AllReserved Usage
}

// For a given resource pool and reservations compute resource utilization per reservation.
// Only capacity groups associated with the given resource pool are considered.
func NewCapacityReservationUsage(snapshot *resourcepool.ResourceSnapshot, reservations []*v1.CapacityGroup) *CapacityReservationUsage {
	inCapacityGroup := map[string]Usage{}
	reserved := Usage{}
	for _, reservation := range reservations {
		if reservation.Spec.ResourcePoolName == snapshot.ResourcePoolName {
			usage := buildUsage(snapshot, reservation)
			inCapacityGroup[reservation.Spec.CapacityGroupName] = usage
			reserved.Allocated = reserved.Allocated.Add(usage.Allocated)
			reserved.Unallocated = reserved.Unallocated.Add(usage.Unallocated)
		}
	}
	return &CapacityReservationUsage{
		InCapacityGroup: inCapacityGroup,
		AllReserved:     reserved,
	}
}

func buildUsage(snapshot *resourcepool.ResourceSnapshot, reservation *v1.CapacityGroup) Usage {
	allocated := v1.ComputeResource{}
	for _, pod := range snapshot.Pods {
		if poolPod.IsPodInCapacityGroup(pod, reservation.Name) && poolPod.IsPodRunning(pod) {
			allocated = allocated.Add(poolPod.FromPodToComputeResource(pod))
		}
	}

	return Usage{
		Allocated:   allocated,
		Unallocated: CapacityGroupResources(reservation).SubWithLimit(allocated, 0),
	}
}
