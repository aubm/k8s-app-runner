//go:generate controller-gen object paths="./..."

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aubm/k8s-app-runner/bare-controller-runtime/api"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
		Complete(&applicationReconcilier{}); err != nil {
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

type applicationReconcilier struct{}

func (r applicationReconcilier) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("received request for pod %s", request.NamespacedName)
	// TODO: do something here
	return reconcile.Result{}, nil
}

type podMutator struct {
	decoder *admission.Decoder
}

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
