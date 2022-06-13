/*
Copyright 2022.

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

// WhitelistSpec defines the desired state of Whitelist
type WhitelistSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Whitelist. Edit whitelist_types.go to remove/update

	Provider  string `json:"provider"`  //外部服务提供商
	Service   string `json:"service"`   //外部服务类型，如rds、slb。。。
	ServiceId string `json:"serviceId"` //

	//+kubebuilder:validation:Enum=Pod;Node;""
	//+kubebuilder:default=Pod
	Level       string            `json:"ipLevel"`               //注册ip的对象(pod\node)
	Annotations map[string]string `json:"annotations,omitempty"` //service对应的参数
}

// WhitelistStatus defines the observed state of Whitelist
type WhitelistStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this
	Created map[string]string `json:"created,omitempty"` // [pod/namespace/name]=registerIp
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".spec.provider",name=Provider,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.service",name=Service,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.serviceId",name=ServiceId,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.ipLevel",name=ipLevel,type=string
//+kubebuilder:resource:scope=Cluster

// Whitelist is the Schema for the whitelists API
type Whitelist struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WhitelistSpec   `json:"spec,omitempty"`
	Status WhitelistStatus `json:"status,omitempty"`
}

func (whl *Whitelist) String() string {
	/*out := fmt.Sprintf("whitelist (%s/%s/%s) for ", whl.Spec.Provider, whl.Spec.Service, whl.Spec.ServiceId)
	for _, k := range whl.Spec.Ownner {
		out = out + "(" + k.Kind + "/" + k.Name + ") "
	}*/
	out := fmt.Sprintf("%+v\n", *whl)
	return out
}

//+kubebuilder:object:root=true

// WhitelistList contains a list of Whitelist
type WhitelistList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Whitelist `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Whitelist{}, &WhitelistList{})
}
