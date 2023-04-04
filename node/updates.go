package node

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonNode "stash.corp.netflix.com/tn/titus-kube-common/node"
	poolUtil "github.com/Netflix/titus-resource-pool/util"
)

func ButNodeCreatedTimestamp(node *v1.Node, timestamp time.Time) *v1.Node {
	node.ObjectMeta.CreationTimestamp = metaV1.Time{Time: timestamp}
	return node
}

func ButNodeLabel(node *v1.Node, key string, value string) *v1.Node {
	if node.Labels == nil {
		node.Labels = map[string]string{}
	}
	node.Labels[key] = value
	return node
}

func ButNodeWithTaint(node *v1.Node, taint *v1.Taint) *v1.Node {
	node.Spec.Taints = append(node.Spec.Taints, *taint)
	return node
}

func ButNodeDecommissioned(source string, node *v1.Node) *v1.Node {
	return ButNodeWithTaint(node, NewDecommissioningTaint(source, time.Now()))
}

func ButNodeScalingDown(source string, node *v1.Node) *v1.Node {
	return ButNodeWithTaint(node, NewScalingDownTaint(source, time.Now()))
}

func ButNodeRemovable(node *v1.Node) *v1.Node {
	node.Labels[commonNode.LabelKeyRemovable] = poolUtil.True
	return node
}

func NewDecommissioningTaint(source string, now time.Time) *v1.Taint {
	return &v1.Taint{
		Key:       commonNode.TaintKeyNodeDecommissioning,
		Value:     source,
		Effect:    "NoExecute",
		TimeAdded: &metaV1.Time{Time: now},
	}
}

func NewScalingDownTaintWithValue(now time.Time, value string) *v1.Taint {
	return &v1.Taint{
		Key:       commonNode.TaintKeyNodeScalingDown,
		Value:     value,
		Effect:    "NoExecute",
		TimeAdded: &metaV1.Time{Time: now},
	}
}

func NewScalingDownTaint(source string, now time.Time) *v1.Taint {
	return NewScalingDownTaintWithValue(now, source)
}
