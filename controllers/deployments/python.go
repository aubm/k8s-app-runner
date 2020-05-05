package deployments

import (
	"fmt"

	k8sapprunnerv1 "github.com/aubm/k8s-app-runner/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Python3DeploymentPropertiesSetter struct{}

func (g *Python3DeploymentPropertiesSetter) CanHandle(app *k8sapprunnerv1.Application) bool {
	return app.Spec.Runtime == "python3"
}

func (g *Python3DeploymentPropertiesSetter) SetDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application) {
	setPythonDeploymentProperties(dep, app, "python:3")
}

type Python2DeploymentPropertiesSetter struct{}

func (g *Python2DeploymentPropertiesSetter) CanHandle(app *k8sapprunnerv1.Application) bool {
	return app.Spec.Runtime == "python2"
}

func (g *Python2DeploymentPropertiesSetter) SetDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application) {
	setPythonDeploymentProperties(dep, app, "python:2")
}

func setPythonDeploymentProperties(dep *appsv1.Deployment, app *k8sapprunnerv1.Application, dockerImage string) {
	setCommonProperties(dep, app)
	volumeMounts := []corev1.VolumeMount{{Name: "source", MountPath: "/opt/app"}}
	dep.Spec.Template.Spec.Volumes = []corev1.Volume{{Name: "source", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}}
	dep.Spec.Template.Spec.InitContainers = []corev1.Container{
		{
			Name:  "download-source",
			Image: "alpine/git:v2.24.3",
			Command: []string{
				"sh",
				"-c",
				fmt.Sprintf("git clone %s /tmp/src && mv /tmp/src/%s /opt/app/src",
					app.Spec.Source.Git.GitRepositoryURL,
					app.Spec.Source.Git.Root,
				),
			},
			VolumeMounts: volumeMounts,
			Resources:    containerResourceRequirements(),
		},
		{
			Name:  "install-dependencies",
			Image: dockerImage,
			Command: []string{
				"bash",
				"-c",
				"pip install virtualenv " +
					"&& virtualenv -p $(which python) /opt/app/venv " +
					"&& [ -f /opt/app/src/requirements.txt ] " +
					"&& source /opt/app/venv/bin/activate " +
					"&& pip install -r /opt/app/src/requirements.txt",
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
				fmt.Sprintf("source /opt/app/venv/bin/activate && python /opt/app/src/%s", app.Spec.Entrypoint),
			},
			VolumeMounts: volumeMounts,
			Ports:        []corev1.ContainerPort{{ContainerPort: app.Spec.Port}},
			Resources:    containerResourceRequirements(),
		},
	}
}
