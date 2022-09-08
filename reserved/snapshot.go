package reserved

import (
	"context"
	"errors"
	"strings"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	capacityGroupV1 "github.com/Netflix/titus-controllers-api/api/capacitygroup/v1"
)

const critical = "critical"

type CapacityGroupSnapshot struct {
	// User provided
	client ctrlClient.Client
	// Loaded
	CapacityGroups       []*capacityGroupV1.CapacityGroup
	CapacityGroupsByName map[string]*capacityGroupV1.CapacityGroup
	// Internal
	capacityGroupByResourcePool map[string][]*capacityGroupV1.CapacityGroup
}

func NewCapacityGroupSnapshot(client ctrlClient.Client) (*CapacityGroupSnapshot, error) {
	snapshot := CapacityGroupSnapshot{
		client: client,
	}

	var err error
	if err = snapshot.ReloadCapacityGroups(); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func NewStaticCapacityGroupSnapshot(capacityGroups []*capacityGroupV1.CapacityGroup) *CapacityGroupSnapshot {
	snapshot := CapacityGroupSnapshot{}
	snapshot.updateCapacityGroupData(capacityGroups)
	return &snapshot
}

func (snapshot *CapacityGroupSnapshot) FindOwnedByResourcePool(resourcePoolName string) []*capacityGroupV1.CapacityGroup {
	return snapshot.capacityGroupByResourcePool[resourcePoolName]
}

func (snapshot *CapacityGroupSnapshot) ReloadCapacityGroups() error {
	if snapshot.client == nil {
		return nil
	}

	capacityGroupList := capacityGroupV1.CapacityGroupList{}
	if err := snapshot.client.List(context.TODO(), &capacityGroupList); err != nil {
		return errors.New("cannot read capacity groups")
	}

	filteredCapacityGroups := filterCapacityGroups(capacityGroupList)
	snapshot.updateCapacityGroupData(filteredCapacityGroups)

	return nil
}

func (snapshot *CapacityGroupSnapshot) updateCapacityGroupData(capacityGroups []*capacityGroupV1.CapacityGroup) {
	capacityGroupByResourcePool := map[string][]*capacityGroupV1.CapacityGroup{}
	capacityGroupsByName := map[string]*capacityGroupV1.CapacityGroup{}
	for _, capacityGroup := range capacityGroups {
		capacityGroupByResourcePool[capacityGroup.Spec.ResourcePoolName] = append(
			capacityGroupByResourcePool[capacityGroup.Spec.ResourcePoolName], capacityGroup)
		capacityGroupsByName[capacityGroup.Spec.CapacityGroupName] = capacityGroup
	}
	snapshot.CapacityGroups = capacityGroups
	snapshot.CapacityGroupsByName = capacityGroupsByName
	snapshot.capacityGroupByResourcePool = capacityGroupByResourcePool
}

// filterCapacityGroups takes in a list of cgs and returns only those that are critical and use kubescheduler
func filterCapacityGroups(cgl capacityGroupV1.CapacityGroupList) []*capacityGroupV1.CapacityGroup {
	var result []*capacityGroupV1.CapacityGroup
	for _, cg := range cgl.Items {
		if isCritical(&cg) && cg.Spec.SchedulerName == PodSchedulerKube {
			result = append(result, cg.DeepCopy())
		}
	}

	return result
}

func isCritical(cg *capacityGroupV1.CapacityGroup) bool {
	if cg.Spec.Tier == "" {
		return true
	}
	return strings.ToLower(cg.Spec.Tier) == critical
}
