package deployments

import (
	"fmt"
	k8sapprunnerv1 "github.com/aubm/k8s-app-runner/kubebuilder/api/v1"
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

func sourceVolume() corev1.Volume {
	return corev1.Volume{Name: "source", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}
}

func sourceVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{Name: "source", MountPath: "/opt/app"}
}

func downloadSourceInitContainer(app *k8sapprunnerv1.Application) corev1.Container {
	return corev1.Container{
		Name:  "download-source",
		Image: "alpine/git:v2.24.3",
		Command: []string{
			"sh",
			"-c",
			fmt.Sprintf("git clone %s /tmp/src && cd /tmp/src && git checkout %s && mv /tmp/src/%s /opt/app/src",
				app.Spec.Source.Git.GitRepositoryURL,
				app.Spec.Source.Git.Revision,
				app.Spec.Source.Git.Root,
			),
		},
		VolumeMounts: []corev1.VolumeMount{sourceVolumeMount()},
		Resources:    containerResourceRequirements(),
	}
}

func envVars(app *k8sapprunnerv1.Application) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{Name: "APP_NAME", Value: app.Name},
		{Name: "APP_RUNTIME", Value: app.Spec.Runtime},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
	}
	for _, envVar := range app.Spec.Env {
		envVars = append(envVars, corev1.EnvVar{Name: envVar.Name, Value: envVar.Value})
	}
	return envVars
}
