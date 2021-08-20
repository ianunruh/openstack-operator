package rally

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func TaskRunnerJob(instance *openstackv1beta1.Rally, keystoneUser *openstackv1beta1.KeystoneUser, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-rally", "/home/rally/.rally/rally.conf", "rally.conf"),
		template.VolumeMount("data", "/var/lib/rally"),
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "runner",
				Image: instance.Spec.Image,
				TTY:   true,
				Command: []string{
					"bash",
					// "-c",
					// template.MustReadFile(AppLabel, "run-task.sh"),
				},
				Env: env,
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(keystoneUser.Spec.Secret),
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "task-runner")

	return job
}

func DBSyncJob(instance *openstackv1beta1.Rally, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-rally", "/home/rally/.rally/rally.conf", "rally.conf"),
	}

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
				Env:          env,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
