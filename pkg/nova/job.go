package nova

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func DBSyncJob(instance *openstackv1beta1.Nova) *batchv1.Job {
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
					template.MustRenderFile(AppLabel, "db-sync.sh", nil),
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("OS_DEFAULT__TRANSPORT_URL", instance.Spec.Broker.Secret, "connection"),
					template.SecretEnvVar("OS_API_DATABASE__CONNECTION", instance.Spec.APIDatabase.Secret, "connection"),
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
					template.SecretEnvVar("CELL_DATABASE", instance.Spec.CellDatabase.Secret, "connection"),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-nova",
						SubPath:   "nova.conf",
						MountPath: "/etc/nova/nova.conf",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-nova", instance.Name, nil),
		},
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
