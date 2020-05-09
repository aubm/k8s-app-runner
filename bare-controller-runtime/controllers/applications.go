package controllers

import (
	"log"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ApplicationReconcilier struct{}

// +kubebuilder:rbac:groups=k8s-app-runner.aubm.net,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s-app-runner.aubm.net,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;patch;update;watch

func (r ApplicationReconcilier) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("received request for pod %s", request.NamespacedName)

	// The bare-controller-runtime implementation actually does nothing.
	// Please consult the implementation with Kubebuilder here:
	// https://github.com/aubm/k8s-app-runner/blob/master/kubebuilder/controllers/application_controller.go

	return reconcile.Result{}, nil
}
