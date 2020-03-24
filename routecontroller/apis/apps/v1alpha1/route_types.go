/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RouteSpec defines the desired state of Route
type RouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Host         string             `json:"host"`
	Path         string             `json:"path,omitempty"`
	Url          string             `json:"url"`
	Domain       RouteDomain        `json:"domain"`
	Destinations []RouteDestination `json:"destinations"`
}

type RouteDomain struct {
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
}

type RouteDestination struct {
	Guid     string              `json:"guid"`
	Weight   *int                `json:"weight,omitempty"`
	Port     *int                `json:"port"`
	App      DestinationApp      `json:"app"`
	Selector DestinationSelector `json:"selector"`
}

type DestinationApp struct {
	Guid    string     `json:"guid"`
	Process AppProcess `json:"process"`
}

type DestinationSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type AppProcess struct {
	Type string `json:"type"`
}

// RouteStatus defines the observed state of Route
type RouteStatus struct {
	Conditions []Condition `json:"conditions"`
}

type Condition struct {
	Type   string `json:"type"`
	Status bool   `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Route is the Schema for the routes API
type Route struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteSpec   `json:"spec,omitempty"`
	Status RouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteList contains a list of Route
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Route `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Route{}, &RouteList{})
}

func (r Route) FQDN() string {
	if r.Spec.Host == "" {
		return r.Spec.Domain.Name
	}
	return fmt.Sprintf("%s.%s", r.Spec.Host, r.Spec.Domain.Name)
}
