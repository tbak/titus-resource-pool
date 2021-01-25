package pod

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sCore "k8s.io/api/core/v1"
)

// We use functions, as K8S records are mutable
var (
	EmptyPod = func() *k8sCore.Pod {
		return &k8sCore.Pod{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "emptyPod",
			},
		}
	}
)
