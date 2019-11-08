package webhook

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BulkSync is an operator-facing configuration for syncing routes from CC into K8s
type BulkSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BulkSyncSpec `json:"spec"`
}

type BulkSyncSpec struct {
	Selector Selector `json:"selector"`
	Template Template `json:"template"`
}

type Selector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type Template struct {
	metav1.ObjectMeta `json:"metadata"`
}

type Service struct {
	ApiVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceSpec `json:"spec"`
}

type ServiceSpec struct {
	Selector map[string]string `json:"selector"`
	Ports    []ServicePort     `json:"ports"`
}

type ServicePort struct {
	Port int    `json:"port"`
	Name string `json:"name"`
}

type HTTPPrefixMatch struct {
	Prefix string `json:"prefix"`
}

type HTTPMatchRequest struct {
	Uri HTTPPrefixMatch `json:"uri"`
}
type VirtualServiceDestination struct {
	Host string `json:"host"`
}
type HTTPRouteDestination struct {
	Destination VirtualServiceDestination `json:"destination"`
	Weight      *int                      `json:"weight,omitempty"`
}

type HTTPRoute struct {
	Match []HTTPMatchRequest     `json:"match,omitempty"`
	Route []HTTPRouteDestination `json:"route,omitempty"`
}

type VirtualServiceSpec struct {
	Hosts    []string    `json:"hosts"`
	Gateways []string    `json:"gateways"`
	Http     []HTTPRoute `json:"http"`
}

type VirtualService struct {
	ApiVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              VirtualServiceSpec `json:"spec"`
}
