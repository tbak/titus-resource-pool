package reserved

import (
	"context"
	"errors"

	v1 "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type CapacityGroupSnapshot struct {
	// User provided
	client ctrlClient.Client
	// Loaded
	CapacityGroups []*v1.CapacityGroup
	// Internal
	capacityGroupByResourcePool map[string][]*v1.CapacityGroup
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

func NewStaticCapacityGroupSnapshot(capacityGroups []*v1.CapacityGroup) *CapacityGroupSnapshot {
	snapshot := CapacityGroupSnapshot{}
	snapshot.updateCapacityGroupData(capacityGroups)
	return &snapshot
}

func (snapshot *CapacityGroupSnapshot) FindOwnByResourcePool(resourcePoolName string) []*v1.CapacityGroup {
	return snapshot.capacityGroupByResourcePool[resourcePoolName]
}

func (snapshot *CapacityGroupSnapshot) ReloadCapacityGroups() error {
	if snapshot.client == nil {
		return nil
	}

	capacityGroupList := v1.CapacityGroupList{}
	if err := snapshot.client.List(context.TODO(), &capacityGroupList); err != nil {
		return errors.New("cannot read capacity groups")
	}
	snapshot.updateCapacityGroupData(AsCapacityGroupReferenceList(&capacityGroupList))
	return nil
}

func (snapshot *CapacityGroupSnapshot) updateCapacityGroupData(capacityGroups []*v1.CapacityGroup) {
	capacityGroupByResourcePool := map[string][]*v1.CapacityGroup{}
	for _, capacityGroup := range capacityGroups {
		capacityGroupByResourcePool[capacityGroup.Spec.ResourcePoolName] = append(
			capacityGroupByResourcePool[capacityGroup.Spec.ResourcePoolName], capacityGroup)
	}
	snapshot.CapacityGroups = capacityGroups
	snapshot.capacityGroupByResourcePool = capacityGroupByResourcePool
}
