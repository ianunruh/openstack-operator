package barbican

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	WorkerComponentLabel = "worker"
)

func WorkerDeployment(instance *openstackv1beta1.Barbican, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, WorkerComponentLabel)

	spec := instance.Spec.Worker

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-barbican", "/etc/barbican/barbican.conf", "barbican.conf"),
		template.SubPathVolumeMount("etc-barbican", "/var/lib/kolla/config_files/config.json", "kolla-barbican-worker.json"),
	}

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)
	pki.AppendRabbitMQTLSClientVolumes(instance.Spec.Broker, &volumes, &volumeMounts)

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:         "worker",
				Image:        spec.Image,
				Command:      []string{"/usr/local/bin/kolla_start"},
				Env:          env,
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "worker")

	return deploy
}
