package node

import (
	"fmt"

	"github.com/google/uuid"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Netflix/titus-kube-common/node"
	"github.com/Netflix/titus-resource-pool/machine"
	"github.com/Netflix/titus-resource-pool/util"

	poolApi "github.com/Netflix/titus-controllers-api/api/resourcepool/v1"
)

const (
	// TODO Use different resource pool name in tests
	ResourcePoolElastic = "elastic"
)

func NewNode(name string, resourcePoolName string, machineTypeConfig *poolApi.MachineTypeConfig) *coreV1.Node {
	return &coreV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				node.LabelKeyResourcePool: resourcePoolName,
			},
		},
		Status: coreV1.NodeStatus{
			Allocatable: util.FromComputeResourceToResourceList(machineTypeConfig.Spec.ComputeResource),
			Capacity:    util.FromComputeResourceToResourceList(machineTypeConfig.Spec.ComputeResource),
		},
	}
}

func NewRandomNode(transformers ...func(node *coreV1.Node)) *coreV1.Node {
	node := NewNode(uuid.New().String()+".node", ResourcePoolElastic, machine.R5Metal())
	for _, transformer := range transformers {
		transformer(node)
	}
	return node
}

func NewNodes(count int64, namePrefix string, resourcePoolName string,
	machineTypeConfig *poolApi.MachineTypeConfig) []*coreV1.Node {
	var nodes []*coreV1.Node
	for i := int64(0); i < count; i++ {
		nodes = append(nodes, NewNode(fmt.Sprintf("%v-%v", namePrefix, i), resourcePoolName,
			machineTypeConfig))
	}
	return nodes
}
