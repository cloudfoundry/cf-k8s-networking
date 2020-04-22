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

package v1alpha3

import (
	"bufio"
	"bytes"

	"github.com/gogo/protobuf/jsonpb"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualServiceSpec defines the desired state of VirtualService
type VirtualServiceSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	istiov1alpha3.VirtualService `json:",inline"`
}

// VirtualServiceStatus defines the observed state of VirtualService
type VirtualServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// VirtualService is the Schema for the virtualservices API
type VirtualService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualServiceSpec   `json:"spec,omitempty"`
	Status VirtualServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualServiceList contains a list of VirtualService
type VirtualServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualService{}, &VirtualServiceList{})
}

func (p *VirtualServiceSpec) MarshalJSON() ([]byte, error) {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)
	marshaler := jsonpb.Marshaler{}
	err := marshaler.Marshal(writer, &p.VirtualService)
	if err != nil {
		return nil, err
	}

	writer.Flush()
	return buffer.Bytes(), nil
}

func (p *VirtualServiceSpec) UnmarshalJSON(b []byte) error {
	reader := bytes.NewReader(b)
	unmarshaler := jsonpb.Unmarshaler{}
	err := unmarshaler.Unmarshal(reader, &p.VirtualService)
	if err != nil {
		return err
	}
	return nil
}
