package neutron

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func DBSyncJob(instance *openstackv1beta1.Neutron, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
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
					template.MustReadFile(AppLabel, "db-sync.sh"),
				},
				Env: env,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-neutron",
						SubPath:   "neutron.conf",
						MountPath: "/etc/neutron/neutron.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
