package reserved

import (
	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
	poolV1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	poolPod "github.com/Netflix/titus-resource-pool/pod"
	"github.com/Netflix/titus-resource-pool/resourcepool"
)

type Usage struct {
	Allocated   poolV1.ComputeResource
	Unallocated poolV1.ComputeResource
}

type CapacityReservationUsage struct {
	// Reservation usage per capacity group.
	InCapacityGroup map[string]Usage
	// Reservation usage for all capacity groups aggregated.
	AllReserved Usage
}

// For a given resource pool and reservations compute resource utilization per reservation.
// Only capacity groups associated with the given resource pool are considered.
func NewCapacityReservationUsage(snapshot *resourcepool.ResourceSnapshot,
	reservations []*capacityGroupV1.CapacityGroup) *CapacityReservationUsage {
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

func buildUsage(snapshot *resourcepool.ResourceSnapshot, reservation *capacityGroupV1.CapacityGroup) Usage {
	allocated := poolV1.ComputeResource{}
	for _, pod := range snapshot.PodSnapshot.ScheduledByName {
		if poolPod.IsPodInCapacityGroup(pod, reservation.Name) {
			allocated = allocated.Add(poolPod.FromPodToComputeResource(pod))
		}
	}

	return Usage{
		Allocated:   allocated,
		Unallocated: CapacityGroupResources(reservation).SubWithLimit(allocated, 0),
	}
}
