module github.com/Netflix/titus-resource-pool

go 1.13

require (
	github.com/Netflix/titus-controllers-api v0.0.11
	github.com/Netflix/titus-kube-common v0.10.1
	github.com/go-logr/logr v0.1.0
	github.com/google/uuid v1.1.1
	github.com/prometheus/client_golang v1.0.0
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/component-base v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)
