package magnum

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ConductorComponentLabel = "conductor"
)

func ConductorStatefulSet(instance *openstackv1beta1.Magnum, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ConductorComponentLabel)

	spec := instance.Spec.Conductor

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-magnum", "/etc/magnum/magnum.conf", "magnum.conf"),
		template.SubPathVolumeMount("etc-magnum", "/var/lib/kolla/config_files/config.json", "kolla-magnum-conductor.json"),
	}

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:         "conductor",
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

	sts.Name = template.Combine(instance.Name, "conductor")

	return sts
}

func ConductorService(instance *openstackv1beta1.Magnum) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ConductorComponentLabel)
	name := template.Combine(instance.Name, "conductor", "headless")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
