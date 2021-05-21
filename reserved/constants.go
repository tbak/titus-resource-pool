package reserved

const (
	PodSchedulerKube  = "kubeScheduler"
	PodSchedulerFenzo = "fenzo"

	bufferCapacityGroupSuffix = "buffer"
)

// GetBufferCapacityGroupName returns a name of a buffer capacity group given the resource pool name.
// A buffer capacity group name format is <resource_pool_name>buffer.
func GetBufferCapacityGroupName(resourcePoolName string) string {
	return resourcePoolName + bufferCapacityGroupSuffix
}
