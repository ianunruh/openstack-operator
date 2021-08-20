package placement

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func DBSyncJob(instance *openstackv1beta1.Placement, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-placement", "/etc/placement/placement.conf", "placement.conf"),
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: instance.Spec.Image,
				Command: []string{
					"placement-manage",
					"db",
					"sync",
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("OS_PLACEMENT_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
