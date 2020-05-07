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

package controllers

import (
	"context"
	"fmt"

	k8sapprunnerv1 "github.com/aubm/k8s-app-runner/kubebuilder/api/v1"
	"github.com/aubm/k8s-app-runner/kubebuilder/controllers/deployments"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
}

// +kubebuilder:rbac:groups=k8s-app-runner.aubm.net,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s-app-runner.aubm.net,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create

func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("application", req.NamespacedName)

	// your logic here
	var app k8sapprunnerv1.Application
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		log.Error(err, "unable to fetch Application")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.reconcileDeployment(ctx, &app); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileHPA(ctx, &app, req); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, &app); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, app *k8sapprunnerv1.Application) error {
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
	res, err := controllerutil.CreateOrUpdate(ctx, r.Client, dep, func() error {
		owner := metav1.GetControllerOf(dep)
		if owner != nil && owner.UID != app.UID {
			err := fmt.Errorf("Found existing deployment which is not controlled by application")
			r.EventRecorder.Event(app, "Warning", "CreateDeployment", err.Error())
			return err
		}
		return r.setDeploymentProperties(dep, app)
	})
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	if res == controllerutil.OperationResultCreated {
		r.EventRecorder.Event(app, "Normal", "CreateDeployment", fmt.Sprintf("Successfully created deployment %s", dep.Name))
	}

	app.Status.AvailableReplicas = dep.Status.AvailableReplicas
	app.Status.Replicas = dep.Status.Replicas
	if err := r.Status().Update(ctx, app); err != nil {
		return err
	}

	return nil
}

func (r *ApplicationReconciler) setDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application) error {
	for _, gen := range deployments.AllDeploymentPropertiesSetters {
		if gen.CanHandle(app) {
			gen.SetDeploymentProperties(dep, app)
		}
	}
	if dep == nil {
		return fmt.Errorf("none of the deployment properties setters could handle the application")
	}

	return ctrl.SetControllerReference(app, dep, r.Scheme)
}

func (r *ApplicationReconciler) reconcileHPA(ctx context.Context, app *k8sapprunnerv1.Application, req ctrl.Request) error {
	hpa := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
	res, err := controllerutil.CreateOrUpdate(ctx, r.Client, hpa, func() error {
		owner := metav1.GetControllerOf(hpa)
		if owner != nil && owner.UID != app.UID {
			err := fmt.Errorf("Found existing hpa which is not controlled by application")
			r.EventRecorder.Event(app, "Warning", "CreateHPA", err.Error())
			return err
		}
		return r.setHPAProperties(hpa, app)
	})
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	if res == controllerutil.OperationResultCreated {
		r.EventRecorder.Event(app, "Normal", "CreateHPA", fmt.Sprintf("Successfully created hpa %s", hpa.Name))
	}

	return nil
}

func (r *ApplicationReconciler) setHPAProperties(hpa *autoscalingv1.HorizontalPodAutoscaler, app *k8sapprunnerv1.Application) error {
	hpa.Labels = map[string]string{"app": app.Name}
	hpa.Name = app.Name
	hpa.Namespace = app.Namespace
	hpa.Spec.MinReplicas = &app.Spec.MinReplicas
	hpa.Spec.MaxReplicas = app.Spec.MaxReplicas
	var targetCPUPercentage int32 = 50
	hpa.Spec.TargetCPUUtilizationPercentage = &targetCPUPercentage
	hpa.Spec.ScaleTargetRef = autoscalingv1.CrossVersionObjectReference{
		Name:       app.Name,
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}
	return ctrl.SetControllerReference(app, hpa, r.Scheme)
}

func (r *ApplicationReconciler) reconcileService(ctx context.Context, app *k8sapprunnerv1.Application) error {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
	res, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		owner := metav1.GetControllerOf(svc)
		if owner != nil && owner.UID != app.UID {
			err := fmt.Errorf("Found existing service which is not controlled by application")
			r.EventRecorder.Event(app, "Warning", "CreateService", err.Error())
			return err
		}
		return r.setServiceProperties(svc, app)
	})
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	if res == controllerutil.OperationResultCreated {
		r.EventRecorder.Event(app, "Normal", "CreateService", fmt.Sprintf("Successfully created service %s", svc.Name))
	}

	ports := make([]int32, 0)
	for _, port := range svc.Spec.Ports {
		if port.NodePort != 0 {
			ports = append(ports, port.NodePort)
		}
	}
	app.Status.NodePort = ports
	if err := r.Status().Update(ctx, app); err != nil {
		return err
	}

	return nil
}

func (r *ApplicationReconciler) setServiceProperties(svc *corev1.Service, app *k8sapprunnerv1.Application) error {
	labels := map[string]string{"app": app.Name}
	svc.Labels = labels
	svc.Name = app.Name
	svc.Namespace = app.Namespace
	svc.Spec.Selector = labels
	svc.Spec.Type = corev1.ServiceTypeNodePort
	if len(svc.Spec.Ports) == 0 {
		svc.Spec.Ports = []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(int(app.Spec.Port))}}
	} else {
		for i, _ := range svc.Spec.Ports {
			svc.Spec.Ports[i].Port = 80
			svc.Spec.Ports[i].TargetPort = intstr.FromInt(int(app.Spec.Port))
		}
	}
	return ctrl.SetControllerReference(app, svc, r.Scheme)
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sapprunnerv1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Complete(r)
}
