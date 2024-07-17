package pki

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func AppendKollaTLSClientVolumes(spec openstackv1beta1.TLSClientSpec, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	AppendTLSClientVolumes(spec, "kolla-ca", "/var/lib/kolla/config_files/ca-certificates/openstack.crt", volumes, volumeMounts)
}

func AppendRabbitMQTLSClientVolumes(spec openstackv1beta1.RabbitMQUserSpec, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	tlsSpec := spec.TLS
	if spec.External != nil {
		tlsSpec = spec.External.TLS
	}
	AppendTLSClientVolumes(tlsSpec, "rabbitmq-ca", "/etc/ssl/certs/rabbitmq/ca.crt", volumes, volumeMounts)
}

func AppendTLSClientVolumes(spec openstackv1beta1.TLSClientSpec, name, mountPath string, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	if spec.CABundle == "" {
		return
	}

	defaultMode := int32(0444)

	*volumes = append(*volumes,
		template.SecretVolume(name, spec.CABundle, &defaultMode))

	*volumeMounts = append(*volumeMounts,
		template.SubPathVolumeMount(name, mountPath, "ca.crt"))
}
