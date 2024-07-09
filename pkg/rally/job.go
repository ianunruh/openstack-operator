package rally

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func DBSyncJob(instance *openstackv1beta1.Rally, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	spec := instance.Spec.DBSyncJob

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-rally", "/home/rally/.rally/rally.conf", "rally.conf"),
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "db-sync.sh"),
				},
				Env: env,
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(template.Combine(instance.Name, "keystone")),
				},
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: spec.NodeSelector,
		Volumes:      volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
