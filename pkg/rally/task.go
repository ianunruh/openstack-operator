package rally

import (
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TaskRunnerJob(instance *openstackv1beta1.RallyTask, cluster *openstackv1beta1.Rally, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

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
					template.MustReadFile(AppLabel, "run-task.sh"),
				},
				Env: env,
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(template.Combine(cluster.Name, "keystone")),
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(cluster.Name, "task", instance.Name)

	return job
}
