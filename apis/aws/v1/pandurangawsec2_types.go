/*
Copyright 2023.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PandurangAWSEC2Spec defines the desired state of PandurangAWSEC2
type PandurangAWSEC2Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PandurangAWSEC2. Edit pandurangawsec2_types.go to remove/update
	Image           string           `json:"image,omitempty"`
	ImagePullPolicy v1.PullPolicy    `json:"imagePullPolicy,omitempty"`
	RestartPolicy   v1.RestartPolicy `json:"restartPolicy,omitempty"`
	Command         string           `json:"command"`
	TagKey          string           `json:"tagKey"`
	TagValue        string           `json:"tagVal"`
	CfgMap          string           `json:"cfgmap"`
}

// PandurangAWSEC2Status defines the observed state of PandurangAWSEC2
type PandurangAWSEC2Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	EC2Status string `json:"ec2status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PandurangAWSEC2 is the Schema for the pandurangawsec2s API
type PandurangAWSEC2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PandurangAWSEC2Spec   `json:"spec,omitempty"`
	Status PandurangAWSEC2Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PandurangAWSEC2List contains a list of PandurangAWSEC2
type PandurangAWSEC2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PandurangAWSEC2 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PandurangAWSEC2{}, &PandurangAWSEC2List{})
}
