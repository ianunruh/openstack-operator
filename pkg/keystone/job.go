package keystone

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func BootstrapJob(instance *openstackv1beta1.Keystone) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "bootstrap",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					"sleep 1",
				},
				Env: []corev1.EnvVar{
					// template.SecretEnvVar("KEYSTONE_ADMIN_PASSWORD", instance.Spec.Secret, "password"),
					// template.EnvVar("KEYSTONE_API_URL", apiURL),
					// template.EnvVar("KEYSTONE_REGION", "RegionOne"),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-keystone",
						SubPath:   "keystone.conf",
						MountPath: "/etc/keystone/keystone.conf",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-keystone", instance.Name, nil),
		},
	})

	job.Name = template.Combine(instance.Name, "bootstrap")

	return job
}

func DBSyncJob(instance *openstackv1beta1.Keystone) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					"sleep 1",
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-keystone",
						SubPath:   "keystone.conf",
						MountPath: "/etc/keystone/keystone.conf",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-keystone", instance.Name, nil),
		},
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
