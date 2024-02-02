package kloudlite

import (
	"golang.org/x/net/context"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kloudlite/internal/client"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kloudlite/internal/constants"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*kloudliteCloudProvider)(nil)

// kloudliteCloudProvider implements cloudprovider.CloudProvider interface.
type kloudliteCloudProvider struct {
	nodeGroups      []cloudprovider.NodeGroup
	KloudliteCli    *client.Client
	ResourceLimiter *cloudprovider.ResourceLimiter
}

func (k *kloudliteCloudProvider) Name() string {
	return "kloudlite"
}

func (k *kloudliteCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	return k.nodeGroups
}

func (k *kloudliteCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// INFO: this node *apiv1.Node is not a complete apiv1.Node object, it does not have labels manytimes
	realNode, err := k.KloudliteCli.GetNode(context.TODO(), node.Name)
	if err != nil {
		return nil, err
	}

	nodepoolName := realNode.Labels[constants.NodepoolNameLabel]
	if nodepoolName == "" {
		return nil, nil
	}

	nodepool, err := k.KloudliteCli.GetNodePoolWithName(context.TODO(), nodepoolName)
	if err != nil {
		return nil, err
	}

	for i := range k.nodeGroups {
		if k.nodeGroups[i].Id() == nodepool.Name {
			return k.nodeGroups[i], nil
		}
	}

	return nil, err
}

func (k *kloudliteCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, nil
}

func (k *kloudliteCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	//TODO implement me
	panic("implement me")
}

func (k *kloudliteCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (k *kloudliteCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (k *kloudliteCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return k.ResourceLimiter, nil
}

func (k *kloudliteCloudProvider) GPULabel() string {
	return constants.GpuLabel
}

func (k *kloudliteCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

func (k *kloudliteCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(k, node)
}

func (k *kloudliteCloudProvider) Cleanup() error {
	return nil
}

func (k *kloudliteCloudProvider) Refresh() error {
	pools, err := k.KloudliteCli.ListNodePools(context.TODO())
	if err != nil {
		return err
	}

	nodeGroups := make([]cloudprovider.NodeGroup, 0, len(pools))

	for i := range pools {
		if pools[i].GetDeletionTimestamp() != nil {
			continue
		}
		nodeGroups = append(nodeGroups, &NodeGroup{
			nodepoolName: pools[i].Name,
			k8sClient:    k.KloudliteCli,
			memoryInGB:   4,
			vcpuCount:    2,
			minSize:      pools[i].Spec.MinCount,
			maxSize:      pools[i].Spec.MaxCount,
		})
	}

	for i := range nodeGroups {
		klog.Infof(nodeGroups[i].Debug())
	}

	k.nodeGroups = nodeGroups
	return nil
}

func BuildKloudlite(opts config.AutoscalingOptions, _ cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	c, err := client.NewClientFromKubeconfigFile(opts.KubeClientOpts.KubeConfigPath)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}

	return &kloudliteCloudProvider{
		KloudliteCli:    c,
		ResourceLimiter: rl,
		nodeGroups:      make([]cloudprovider.NodeGroup, 0),
	}
}
