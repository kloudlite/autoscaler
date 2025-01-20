package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kloudlite/internal/constants"
	t "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kloudlite/internal/types"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type Client struct {
	k8sCli client.Client
}

func jsonConvert(from any, to any) error {
	b, err := json.Marshal(from)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, to)
}

func (k *Client) getNodePool(ctx context.Context, name string) (*t.NodePool, error) {
	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": t.NodePoolGVK.GroupVersion().String(),
			"kind":       t.NodePoolGVK.Kind,
		},
	}
	err := k.k8sCli.Get(ctx, types.NamespacedName{Name: name, Namespace: ""}, obj)
	if err != nil {
		return nil, err
	}

	var nodepool t.NodePool
	if err := jsonConvert(obj.Object, &nodepool); err != nil {
		return nil, err
	}

	return &nodepool, nil
}

func (k *Client) UpdateNodepoolTargetSize(ctx context.Context, name string, targetSize int) error {
	obj := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": t.NodePoolGVK.GroupVersion().String(),
			"kind":       t.NodePoolGVK.Kind,
		},
	}

	if err := k.k8sCli.Get(ctx, types.NamespacedName{Name: name}, &obj); err != nil {
		return err
	}

	// INFO: this is a hack to make sure the nodepool controller will not reconcile this nodepool in the next 10 seconds
	metadata := obj.Object["metadata"].(map[string]any)
	if metadata["annotations"] == nil {
		metadata["annotations"] = make(map[string]any)
	}
	annotations := metadata["annotations"].(map[string]any)
	annotations[constants.AnnotationReconcileScheduledAfter] = time.Now().Add(10 * time.Second).Format(time.RFC3339)
	return k.k8sCli.Update(ctx, &obj)
}

func (k *Client) listNodePools(ctx context.Context) ([]*t.NodePool, error) {
	obj := &unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": t.NodePoolGVK.GroupVersion().String(),
			"kind":       t.NodePoolGVK.Kind,
		},
	}

	if err := k.k8sCli.List(ctx, obj); err != nil {
		return nil, err
	}

	var nodepools []*t.NodePool
	if err := jsonConvert(obj.Items, &nodepools); err != nil {
		return nil, err
	}
	return nodepools, nil
}

func (k *Client) getNode(ctx context.Context, nodeName string) (*corev1.Node, error) {
	var node corev1.Node
	if err := k.k8sCli.Get(ctx, types.NamespacedName{Name: nodeName}, &node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (k *Client) ListNodes(ctx context.Context, poolName string) ([]corev1.Node, error) {
	nodesList := unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "clusters.kloudlite.io/v1",
			"kind":       "Node",
		},
	}

	nodes := make([]corev1.Node, len(nodesList.Items))
	for i := range nodesList.Items {
		n, err := k.getNode(ctx, nodesList.Items[i].GetName())
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return nil, err
			}
			n = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nodesList.Items[i].GetName(),
					Namespace: nodesList.Items[i].GetNamespace(),
				},
			}
		}
		nodes[i] = *n
	}

	return nodes, nil
}

func (k *Client) TargetSize(nodepoolName string) (int, error) {
	list := unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "clusters.kloudlite.io/v1",
			"kind":       "Node",
		},
		Items: nil,
	}
	if err := k.k8sCli.List(context.TODO(), &list, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodepoolNameLabel: nodepoolName,
		}),
	}); err != nil {
		return 0, err
	}

	count := 0
	for i := range list.Items {
		if list.Items[i].GetDeletionTimestamp() == nil {
			count += 1
		}
	}

	return count, nil
}

func (k *Client) GetNodePoolWithName(ctx context.Context, name string) (*t.NodePool, error) {
	return k.getNodePool(ctx, name)
}

func (k *Client) ListNodePools(ctx context.Context) ([]*t.NodePool, error) {
	return k.listNodePools(ctx)
}

func (k *Client) GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	return k.getNode(ctx, name)
}

func (k *Client) DeleteNode(ctx context.Context, name string) error {
	customNode := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": t.NodeGVK.GroupVersion().String(),
			"kind":       t.NodeGVK.Kind,
			"metadata": map[string]any{
				"name": name,
			},
		},
	}

	if err := k.k8sCli.Delete(ctx, &customNode); err != nil {
		return err
	}

	klog.Infof("Node %s, has been marked for deletion, nodepool controller should take it forward, now", name)
	return nil
}

func (k *Client) CreateNode(ctx context.Context, nodepoolName string) error {
	obj := map[string]any{
		"apiVersion": "clusters.kloudlite.io/v1",
		"kind":       "Node",
		"metadata": map[string]any{
			"generateName": fmt.Sprintf("%s-node-", nodepoolName),
			"labels": map[string]string{
				"kloudlite.io/nodepool.name": nodepoolName,
			},
		},
		"spec": map[string]any{
			"nodepoolName": nodepoolName,
		},
	}
	return k.k8sCli.Create(ctx, &unstructured.Unstructured{Object: obj})
}

func (k *Client) DeleteNodes(ctx context.Context, nodes []*corev1.Node) error {
	for i := range nodes {
		node := nodes[i]
		return k.DeleteNode(ctx, node.Name)
	}

	return nil
}

func (k *Client) InstanceStatusFromNode(node *corev1.Node) *cloudprovider.InstanceStatus {
	return toInstanceStatus(node)
}

func toInstanceState(node *corev1.Node) (cloudprovider.InstanceState, error) {
	if node.GetDeletionTimestamp() != nil {
		return cloudprovider.InstanceDeleting, nil
	}

	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == corev1.NodeReady {
			return cloudprovider.InstanceRunning, nil
		}
	}

	return cloudprovider.InstanceState(-1), fmt.Errorf("unknown instance state")
}

// referenced from hetzner's cluster-autoscaler/cloudprovider/hetzner/hetzner_node_group.go:292
func toInstanceStatus(node *corev1.Node) *cloudprovider.InstanceStatus {
	st := &cloudprovider.InstanceStatus{}

	state, err := toInstanceState(node)
	if err != nil {
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-kloudlite",
			ErrorMessage: err.Error(),
		}
		return st
	}

	st.State = state

	return st
}

func NewClientFromKubeconfigFile(kubeconfigPath string) (*Client, error) {
	klog.V(1).Infof("Using kubeconfig file: %s", kubeconfigPath)
	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		klog.Fatalf("Failed to parse kubeconfig file: %v", err)
	}

	return NewClient(restCfg)
}

func NewClient(config *rest.Config) (*Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientGoScheme.AddToScheme(scheme))

	c, err := client.New(config, client.Options{
		Scheme: scheme,
		WarningHandler: client.WarningHandlerOptions{
			SuppressWarnings: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &Client{k8sCli: c}, nil
}
