package kloudlite

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ cloudprovider.CloudProvider = (*kloudliteCloudProvider)(nil)

// kloudliteCloudProvider implements CloudProvider interface.
type kloudliteCloudProvider struct {
	k8sCli client.Client
}

func (k kloudliteCloudProvider) Name() string {
	return "kloudlite"
}

func (k kloudliteCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	//TODO implement me
	panic("implement me, nodegroups")
}

func (k kloudliteCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) GPULabel() string {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) Cleanup() error {
	//TODO implement me
	panic("implement me")
}

func (k kloudliteCloudProvider) Refresh() error {
	//TODO implement me
	panic("implement me")
}

func BuildKloudlite(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	return &kloudliteCloudProvider{}
}
