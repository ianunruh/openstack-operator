package glance

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func DBSyncJob(instance *openstackv1beta1.Glance) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: instance.Spec.Image,
				Command: []string{
					"glance-manage",
					"db_sync",
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-glance",
						SubPath:   "glance-api.conf",
						MountPath: "/etc/glance/glance-api.conf",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-glance", instance.Name, nil),
		},
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
