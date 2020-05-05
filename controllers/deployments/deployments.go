package deployments

import (
	k8sapprunnerv1 "github.com/aubm/k8s-app-runner/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AllDeploymentPropertiesSetters = []DeploymentPropertiesSetter{
	&Python2DeploymentPropertiesSetter{},
	&Python3DeploymentPropertiesSetter{},
	&Node12DeploymentPropertiesSetter{},
	&Node14DeploymentPropertiesSetter{},
}

type DeploymentPropertiesSetter interface {
	CanHandle(app *k8sapprunnerv1.Application) bool
	SetDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application)
}

func setCommonProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application) {
	labels := map[string]string{"app": app.Name}
	dep.Labels = labels
	dep.Name = app.Name
	dep.Namespace = app.Namespace
	dep.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	dep.Spec.Template.ObjectMeta.Labels = labels
	if dep.Spec.Replicas == nil {
		dep.Spec.Replicas = &app.Spec.MinReplicas
	}
}

func containerResourceRequirements() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("120Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
}
