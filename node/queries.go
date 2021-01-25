package node

import k8sCore "k8s.io/api/core/v1"

func AsNodeReferenceList(nodeList *k8sCore.NodeList) []*k8sCore.Node {
	result := []*k8sCore.Node{}
	for _, node := range nodeList.Items {
		tmp := node
		result = append(result, &tmp)
	}
	return result
}
