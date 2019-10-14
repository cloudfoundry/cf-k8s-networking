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

// Route represents a CC v3 Route
type Route struct {
	ApiVersion        string `json:"apiVersion"`
	Kind              string `json:"kind"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              RouteSpec `json:"spec"`
}

type RouteSpec struct {
	Selector     Selector      `json:"selector"`
	Host         string        `json:"host"`
	Path         string        `json:"path"`
	Domain       Domain        `json:"domain"`
	Destinations []Destination `json:"destinations"`
}

type Domain struct {
	Guid     string `json:"guid"`
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
}

type Destination struct {
	Guid   string `json:"guid"`
	Port   int    `json:"port"`
	Weight *int   `json:"weight,omitempty"`
	App    App    `json:"app"`
}

type App struct {
	Guid    string  `json:"guid"`
	Process Process `json:"process"`
}

type Process struct {
	Type string `json:"type"`
}
