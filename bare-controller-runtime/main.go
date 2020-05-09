//go:generate controller-gen object rbac:roleName=manager-role webhook paths="./..."

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aubm/k8s-app-runner/bare-controller-runtime/api"
	"github.com/aubm/k8s-app-runner/bare-controller-runtime/controllers"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func main() {
	if err := _main(); err != nil {
		log.Fatal(err.Error())
	}
}

func _main() error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Port: 9443})
	if err != nil {
		return fmt.Errorf("failed to create new manager: %v", err)
	}

	if err := api.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("failed to add custom types to scheme: %v", err)
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&api.Application{}).
		Complete(&controllers.ApplicationReconcilier{}); err != nil {
		return fmt.Errorf("failed to create new controller: %v", err)
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&api.Application{}).
		Complete(); err != nil {
		return fmt.Errorf("failed to create webhooks for Application")
	}

	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/mutate-pod", &webhook.Admission{Handler: &podMutator{}})

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("failed to start manager: %v", err)
	}

	return nil
}

type podMutator struct {
	decoder *admission.Decoder
}

// +kubebuilder:webhook:verbs=create;update,path=/mutate-pod,mutating=true,failurePolicy=fail,groups=core,resources=pods,versions=v1,name=mpod.kb.io

func (m *podMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := m.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations["from-pod-mutator-with-love"] = "Hello there"

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (m *podMutator) InjectDecoder(d *admission.Decoder) error {
	m.decoder = d
	return nil
}
