package nova

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

func ConductorStatefulSet(name, namespace string, spec openstackv1beta1.NovaConductorSpec, tlsSpec openstackv1beta1.TLSClientSpec, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)

	probe := &corev1.Probe{
		ProbeHandler:        amqpHealthProbeHandler("nova-conductor"),
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
		template.SubPathVolumeMount("etc-nova", "/var/lib/kolla/config_files/config.json", "kolla-nova-conductor.json"),
	}

	pki.AppendTLSClientVolumes(tlsSpec, &volumes, &volumeMounts)

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:          "conductor",
				Image:         spec.Image,
				Command:       []string{"/usr/local/bin/kolla_start"},
				Env:           env,
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     spec.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	sts.Name = template.Combine(name, "conductor")

	return sts
}

func ConductorService(name, namespace string) *corev1.Service {
	labels := template.Labels(name, AppLabel, ConductorComponentLabel)
	name = template.Combine(name, "conductor")

	svc := template.GenericService(name, namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}
