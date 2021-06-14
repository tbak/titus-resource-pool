module github.com/Netflix/titus-resource-pool

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.19.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.9
	k8s.io/client-go => k8s.io/client-go v0.19.9
	k8s.io/component-base => k8s.io/component-base v0.19.9
)

require (
	github.com/Netflix/titus-controllers-api v0.0.13
	github.com/Netflix/titus-kube-common v0.15.0
	github.com/go-logr/logr v0.3.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.1.2
	github.com/prometheus/client_golang v1.7.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/component-base v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/structured-merge-diff/v4 v4.0.3 // indirect
)
