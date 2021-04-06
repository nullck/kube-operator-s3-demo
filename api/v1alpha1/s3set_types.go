/*
Copyright 2021.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S3SetSpec defines the desired state of S3Set
type S3SetSpec struct {
	BucketName string `json:"bucketname,omitempty"`
}

// S3SetStatus defines the observed state of S3Set
type S3SetStatus struct {
}

// S3Set is the Schema for the s3sets API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
type S3Set struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3SetSpec   `json:"spec,omitempty"`
	Status S3SetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// S3SetList contains a list of S3Set
type S3SetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Set `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3Set{}, &S3SetList{})
}
