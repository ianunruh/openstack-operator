package pki

import (
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
	corev1 "k8s.io/api/core/v1"
)

func AppendTLSServerVolumes(spec openstackv1beta1.TLSServerSpec, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	if spec.Secret == "" {
		return
	}

	defaultMode := int32(0400)

	*volumes = append(*volumes,
		template.SecretVolume("secret-tls", spec.Secret, &defaultMode))

	*volumeMounts = append(*volumeMounts,
		template.VolumeMount("secret-tls", "/etc/keystone/certs"))
}

func HTTPActionScheme(spec openstackv1beta1.TLSServerSpec) corev1.URIScheme {
	if spec.Secret == "" {
		return corev1.URISchemeHTTP
	}
	return corev1.URISchemeHTTPS
}
