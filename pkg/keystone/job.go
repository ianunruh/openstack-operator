package keystone

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func BootstrapJob(instance *openstackv1beta1.Keystone) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	apiURL := fmt.Sprintf("https://%s/v3", instance.Spec.API.Ingress.Host)
	apiInternalURL := fmt.Sprintf("http://%s-api.%s.svc:5000/v3", instance.Name, instance.Namespace)

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
					template.MustRenderFile(AppLabel, "bootstrap.sh", nil),
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
					template.SecretEnvVar("KEYSTONE_ADMIN_PASSWORD", instance.Name, "OS_PASSWORD"),
					template.EnvVar("KEYSTONE_API_URL", apiURL),
					template.EnvVar("KEYSTONE_API_INTERNAL_URL", apiInternalURL),
					template.EnvVar("KEYSTONE_REGION", "RegionOne"),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-keystone",
						SubPath:   "keystone.conf",
						MountPath: "/etc/keystone/keystone.conf",
					},
					{
						Name:      "fernet-keys",
						MountPath: "/etc/keystone/fernet-keys",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-keystone", instance.Name, nil),
			template.SecretVolume("fernet-keys", template.Combine(instance.Name, "fernet-keys"), nil),
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
					"keystone-manage",
					"db_sync",
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
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
