package keystone

import (
	"slices"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func BootstrapJob(instance *openstackv1beta1.Keystone, env []corev1.EnvVar, volumes []corev1.Volume) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	spec := instance.Spec.BootstrapJob

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-keystone", "/etc/keystone/keystone.conf", "keystone.conf"),
		template.VolumeMount("fernet-keys", "/etc/keystone/fernet-keys"),
	}

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

	env = slices.Concat(env, []corev1.EnvVar{
		template.SecretEnvVar("KEYSTONE_ADMIN_PASSWORD", instance.Name, "OS_PASSWORD"),
		template.EnvVar("KEYSTONE_API_URL", APIPublicURL(instance)),
		template.EnvVar("KEYSTONE_API_INTERNAL_URL", APIInternalURL(instance)),
		template.EnvVar("KEYSTONE_REGION", "RegionOne"),
	})

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "bootstrap",
				Image: instance.Spec.API.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "bootstrap.sh"),
				},
				Env:          env,
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: spec.NodeSelector,
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

	spec := instance.Spec.DBSyncJob

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-keystone", "/etc/keystone/keystone.conf", "keystone.conf"),
	}

	pki.AppendKollaTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: instance.Spec.API.Image,
				Command: []string{
					"keystone-manage",
					"db_sync",
				},
				Env:          env,
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: spec.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}
