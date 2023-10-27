package kloudlite

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kloudlite/internal/client"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kloudlite/internal/constants"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"math"
)

// NodeGroup implements cloudprovider.NodeGroup interface.
type NodeGroup struct {
	nodepoolName string

	k8sClient *client.Client

	memoryInGB float32
	vcpuCount  int64

	minSize    int
	maxSize    int
	targetSize int
}

var _ cloudprovider.NodeGroup = (*NodeGroup)(nil)

func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

func (n *NodeGroup) MinSize() int {
	return n.minSize
}

func (n *NodeGroup) TargetSize() (int, error) {
	return n.targetSize, nil
}

func (n *NodeGroup) IncreaseSize(delta int) error {
	if err := n.k8sClient.UpdateNodepoolTargetSize(context.TODO(), n.nodepoolName, n.targetSize+delta); err != nil {
		return err
	}
	n.targetSize += delta
	return nil
}

func (n *NodeGroup) DeleteNodes(nodes []*corev1.Node) error {
	if err := n.k8sClient.UpdateNodepoolTargetSize(context.TODO(), n.nodepoolName, int(math.Max(float64(n.targetSize-len(nodes)), float64(n.minSize)))); err != nil {
		return err
	}
	if err := n.k8sClient.DeleteNodes(context.TODO(), nodes); err != nil {
		return err
	}
	return nil
}

func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if err := n.k8sClient.UpdateNodepoolTargetSize(context.TODO(), n.nodepoolName, n.targetSize-delta); err != nil {
		return err
	}
	n.targetSize -= delta
	return nil
}

func (n *NodeGroup) Id() string {
	return n.nodepoolName
}

func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("[DEBUG]: nodepoolName: %s, minSize: %d, maxSize: %d, targetSize: %d", n.nodepoolName, n.minSize, n.maxSize, n.targetSize)
}

func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := n.k8sClient.ListNodes(context.TODO(), n.nodepoolName)
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		klog.Infof("Node.Name: %v", nodes[i].Name)
	}

	instances := make([]cloudprovider.Instance, len(nodes))
	for i := range nodes {
		instances[i] = cloudprovider.Instance{
			Id:     nodes[i].Name,
			Status: n.k8sClient.InstanceStatusFromNode(&nodes[i]),
		}
	}
	return instances, nil
}

func (n *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	nodeInfo := schedulerframework.NewNodeInfo()

	templateNode := &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "template-node",
			Labels: map[string]string{constants.NodepoolNameLabel: n.nodepoolName},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
				corev1.ResourceCPU:     *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceMemory:  *resource.NewQuantity(2*1024*1024*1024, resource.DecimalSI),
				corev1.ResourceStorage: *resource.NewQuantity(30*1024*1024*1024, resource.DecimalSI),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
				corev1.ResourceCPU:     *resource.NewQuantity(2, resource.DecimalSI),
				corev1.ResourceMemory:  *resource.NewQuantity(2*1024*1024*1024, resource.DecimalSI),
				corev1.ResourceStorage: *resource.NewQuantity(30*1024*1024*1024, resource.DecimalSI),
			},
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}

	nodeInfo.SetNode(templateNode)
	return nodeInfo, nil
}

func (n *NodeGroup) Exist() bool {
	return true
}

func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NodeGroup) Delete() error {
	panic("implement me")
}

func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return &defaults, nil
}
