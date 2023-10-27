package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NodePoolSpec struct {
	MaxCount    int `json:"maxCount"`
	MinCount    int `json:"minCount"`
	TargetCount int `json:"targetCount"`
}

// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NodePoolSpec `json:"spec,omitempty"`
}

var NodePoolGVK = schema.GroupVersionKind{
	Group:   "clusters.kloudlite.io",
	Version: "v1",
	Kind:    "NodePool",
}

var NodeGVK = schema.GroupVersionKind{
	Group:   "clusters.kloudlite.io",
	Version: "v1",
	Kind:    "Node",
}

func (np *NodePool) EnsureGVK() {
	if np != nil {
		np.SetGroupVersionKind(NodePoolGVK)
	}
}

type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}
