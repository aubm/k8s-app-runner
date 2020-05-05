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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// The application definition
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Port which the application listen
	// +kubebuilder:default=8080
	// +optional
	Port int32 `json:"port"`

	// Application runtime name and version
	// +kubebuilder:validation:Enum=python2;python3;node12;node14
	Runtime string `json:"runtime"`

	// Minimum number of replicas for the application
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	// +optional
	MinReplicas int32 `json:"minReplicas"`

	// Maximum number of replicas for the application
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10
	// +optional
	MaxReplicas int32 `json:"maxReplicas"`

	// The program entrypoint, for example "main.py"
	// +kubebuilder:validation:MinLength=1
	Entrypoint string `json:"entrypoint"`

	Source ApplicationSource `json:"source"`
}

type ApplicationSource struct {
	Git ApplicationSourceGit `json:"git"`
}

type ApplicationSourceGit struct {
	// Git repository URL for fetching sources
	// +kubebuilder:validation:MinLength=1
	GitRepositoryURL string `json:"gitRepositoryUrl"`

	// Root folder of the application source in the repository tree, defaults to "" (repository's root)
	// +kubebuilder:validation:MinLength=1
	// +optional
	Root string `json:"root,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Node port on which the deployed application listen
	// +optional
	NodePort []int32 `json:"nodePort,omitempty"`

	// Total number of non-terminated pods
	// +optional
	AvailableReplicas int32 `json:"availableReplicas"`

	// Total number of non-terminated pods targeted
	// +optional
	Replicas int32 `json:"replicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="RUNTIME",type=string,JSONPath=`.spec.runtime`
// +kubebuilder:printcolumn:name="NODE PORT",type=number,JSONPath=`.status.nodePort[0]`
// +kubebuilder:printcolumn:name="TARGETED REPLICAS",type=number,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="AVAILABLE REPLICAS",type=number,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
