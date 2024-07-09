package tlsproxy

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func MustReadConfig() string {
	return template.MustReadFile("tlsproxy", "haproxy.cfg")
}

func Container(port int32, spec openstackv1beta1.TLSProxySpec, probe *corev1.Probe, volumeMounts []corev1.VolumeMount) corev1.Container {
	runAsUser := int64(0)

	return corev1.Container{
		Name:  "tlsproxy",
		Image: spec.Image,
		Command: []string{
			"bash",
			"-c",
			template.MustReadFile("tlsproxy", "run.sh"),
		},
		Env: []corev1.EnvVar{
			template.EnvVar("SERVICE_HTTP_PORT", strconv.Itoa(int(port))),
			template.FieldEnvVar("SERVICE_BIND_IP", "status.podIP"),
		},
		Ports: []corev1.ContainerPort{
			{Name: "http", ContainerPort: port},
		},
		Resources: spec.Resources,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  &runAsUser,
			RunAsGroup: &runAsUser,
		},
		VolumeMounts: volumeMounts,
	}
}

func VolumeMounts(name, subPath string) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		template.SubPathVolumeMount(name, "/usr/local/etc/haproxy/haproxy.cfg", subPath),
	}
}

func AppendTLSServerVolumes(spec openstackv1beta1.TLSServerSpec, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	pki.AppendTLSServerVolumes(spec, "/usr/local/etc/haproxy/certs", 0400, volumes, volumeMounts)
}
