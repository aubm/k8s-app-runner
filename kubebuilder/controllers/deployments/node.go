package deployments

import (
	"fmt"

	k8sapprunnerv1 "github.com/aubm/k8s-app-runner/kubebuilder/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Node12DeploymentPropertiesSetter struct{}

func (g *Node12DeploymentPropertiesSetter) CanHandle(app *k8sapprunnerv1.Application) bool {
	return app.Spec.Runtime == "node12"
}

func (g *Node12DeploymentPropertiesSetter) SetDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application) {
	setNodeDeploymentProperties(dep, app, "node:12.16.3")
}

type Node14DeploymentPropertiesSetter struct{}

func (g *Node14DeploymentPropertiesSetter) CanHandle(app *k8sapprunnerv1.Application) bool {
	return app.Spec.Runtime == "node14"
}

func (g *Node14DeploymentPropertiesSetter) SetDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application) {
	setNodeDeploymentProperties(dep, app, "node:14.1.0")
}

func setNodeDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application, dockerImage string) {
	setCommonProperties(dep, app)
	volumeMounts := []corev1.VolumeMount{sourceVolumeMount()}
	dep.Spec.Template.Spec.Volumes = []corev1.Volume{sourceVolume()}
	dep.Spec.Template.Spec.InitContainers = []corev1.Container{
		downloadSourceInitContainer(app),
		{
			Name:  "install-dependencies",
			Image: dockerImage,
			Command: []string{
				"bash",
				"-c",
				"[ -f /opt/app/src/package.json ] && cd /opt/app/src && npm install",
			},
			VolumeMounts: volumeMounts,
			Resources:    containerResourceRequirements(),
		},
	}
	dep.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "app",
			Image: dockerImage,
			Command: []string{
				"bash",
				"-c",
				fmt.Sprintf("node /opt/app/src/%s", app.Spec.Entrypoint),
			},
			VolumeMounts: volumeMounts,
			Env:          envVars(app),
			Ports:        []corev1.ContainerPort{{ContainerPort: app.Spec.Port}},
			Resources:    containerResourceRequirements(),
		},
	}
}
