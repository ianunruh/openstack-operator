package amphora

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Volumes(instance *openstackv1beta1.Octavia) []corev1.Volume {
	defaultMode := int32(0400)
	return []corev1.Volume{
		template.SecretVolume("cert-client",
			template.Combine(instance.Name, "amphora-client"), &defaultMode),
		template.EmptyDirVolume("cert-client-combined"),
		template.SecretVolume("cert-server-ca",
			template.Combine(instance.Name, "amphora-server-ca-root"), &defaultMode),
	}
}

func VolumeMounts(instance *openstackv1beta1.Octavia) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		template.VolumeMount("cert-client", "/etc/octavia/certs/client"),
		template.VolumeMount("cert-client-combined", "/etc/octavia/certs/client-combined"),
		template.VolumeMount("cert-server-ca", "/etc/octavia/certs/server-ca"),
	}
}

func InitContainer(image string, volumeMounts []corev1.VolumeMount) corev1.Container {
	return corev1.Container{
		Name:  "init-pki",
		Image: image,
		Command: []string{
			"bash",
			"-c",
			template.MustReadFile("octavia", "init-pki.sh"),
		},
		VolumeMounts: volumeMounts,
	}
}
