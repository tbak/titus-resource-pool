module github.com/Netflix/titus-resource-pool

go 1.15

replace (
	k8s.io/api => k8s.io/api v0.20.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.9
	k8s.io/client-go => k8s.io/client-go v0.20.9
	k8s.io/component-base => k8s.io/component-base v0.20.9
)

require (
	github.com/Netflix/titus-controllers-api v0.0.13
	github.com/Netflix/titus-kube-common v0.20.0
	github.com/go-logr/logr v0.3.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golangci/golangci-lint v1.30.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/prometheus/client_golang v1.7.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.9
	k8s.io/apimachinery v0.20.9
	k8s.io/component-base v0.20.9
	sigs.k8s.io/controller-runtime v0.7.2
)
