package task

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rally"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func RunnerJob(instance *openstackv1beta1.RallyTask, cluster *openstackv1beta1.Rally, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, rally.AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-rally", "/home/rally/.rally/rally.conf", "rally.conf"),
		// template.VolumeMount("data", "/var/lib/rally"),
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "runner",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(rally.AppLabel, "run-task.sh"),
				},
				Env:          env,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(cluster.Name, "task", instance.Name)

	return job
}
