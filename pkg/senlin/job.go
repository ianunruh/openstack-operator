package senlin

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func DBSyncJob(instance *openstackv1beta1.Senlin, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-senlin", "/etc/senlin/senlin.conf", "senlin.conf"),
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: instance.Spec.Image,
				Command: []string{
					"senlin-manage",
					"db_sync",
				},
				Env:          env,
				Resources:    instance.Spec.DBSyncJob.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: instance.Spec.DBSyncJob.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
