package keystone

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func BootstrapJob(instance *openstackv1beta1.Keystone, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	apiURL := fmt.Sprintf("https://%s/v3", instance.Spec.API.Ingress.Host)
	apiInternalURL := fmt.Sprintf("http://%s-api.%s.svc:5000/v3", instance.Name, instance.Namespace)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-keystone", "/etc/keystone/keystone.conf", "keystone.conf"),
		template.VolumeMount("fernet-keys", "/etc/keystone/fernet-keys"),
	}

	extraEnv := []corev1.EnvVar{
		template.SecretEnvVar("KEYSTONE_ADMIN_PASSWORD", instance.Name, "OS_PASSWORD"),
		template.EnvVar("KEYSTONE_API_URL", apiURL),
		template.EnvVar("KEYSTONE_API_INTERNAL_URL", apiInternalURL),
		template.EnvVar("KEYSTONE_REGION", "RegionOne"),
	}

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
					template.MustReadFile(AppLabel, "bootstrap.sh"),
				},
				Env:          append(env, extraEnv...),
				Resources:    instance.Spec.BootstrapJob.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: instance.Spec.BootstrapJob.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "bootstrap")

	return job
}

func DBSyncJob(instance *openstackv1beta1.Keystone, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-keystone", "/etc/keystone/keystone.conf", "keystone.conf"),
	}

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
