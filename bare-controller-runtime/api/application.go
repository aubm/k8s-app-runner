package api

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	SchemeBuilder = &scheme.Builder{GroupVersion: schema.GroupVersion{
		Group:   "k8s-app-runner.aubm.net",
		Version: "v1",
	}}
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}

// +kubebuilder:object:root=true
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {
	Port        int32  `json:"port"`
	Runtime     string `json:"runtime"`
	MinReplicas int32  `json:"minReplicas"`
	MaxReplicas int32  `json:"maxReplicas"`
	Env         []Env  `json:"env"`
	Entrypoint  string `json:"entrypoint"`
	Source      struct {
		Git struct {
			GitRepositoryURL string `json:"gitRepositoryUrl"`
			Revision         string `json:"revision"`
			Root             string `json:"root,omitempty"`
		} `json:"git"`
	} `json:"source"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ApplicationStatus struct {
	NodePort          []int32 `json:"nodePort,omitempty"`
	AvailableReplicas int32   `json:"availableReplicas"`
	Replicas          int32   `json:"replicas"`
}

// +kubebuilder:object:root=true
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

var _ webhook.Defaulter = &Application{}

func (in *Application) Default() {
	if in.ObjectMeta.Annotations == nil {
		in.ObjectMeta.Annotations = map[string]string{}
	}
	in.ObjectMeta.Annotations["from-application-mutator-with-love"] = "Hello there"
}

var _ webhook.Validator = &Application{}

func (in *Application) ValidateCreate() error {
	return in.validate()
}

func (in *Application) ValidateUpdate(old runtime.Object) error {
	return in.validate()
}

func (in *Application) ValidateDelete() error {
	return nil
}

func (in *Application) validate() error {
	if in.Spec.MinReplicas > in.Spec.MaxReplicas {
		return fmt.Errorf("minReplicas can not be greater than maxReplicas, minReplicas=%v maxReplicas=%v",
			in.Spec.MinReplicas, in.Spec.MaxReplicas)
	}
	return nil
}
