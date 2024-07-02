package pki

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func AppendTLSClientVolumes(spec openstackv1beta1.TLSClientSpec, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	if spec.CABundle == "" {
		return
	}

	defaultMode := int32(0444)

	*volumes = append(*volumes,
		template.SecretVolume("tls-ca-bundle", spec.CABundle, &defaultMode))

	*volumeMounts = append(*volumeMounts,
		template.SubPathVolumeMount("tls-ca-bundle", "/var/lib/kolla/config_files/ca-certificates/openstack.crt", "ca.crt"))
}
