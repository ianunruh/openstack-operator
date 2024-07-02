package pki

import (
	"gopkg.in/ini.v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func SetupKeystoneMiddleware(cfg *ini.File, spec openstackv1beta1.TLSClientSpec) {
	if spec.CABundle != "" {
		cfg.Section("keystone_authtoken").NewKey("cafile", "/etc/ssl/certs/openstack-ca.crt")
	}
}

func AppendTLSClientVolumes(spec openstackv1beta1.TLSClientSpec, volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) {
	if spec.CABundle == "" {
		return
	}

	defaultMode := int32(0444)

	*volumes = append(*volumes,
		template.SecretVolume("tls-ca-bundle", spec.CABundle, &defaultMode))

	*volumeMounts = append(*volumeMounts,
		template.SubPathVolumeMount("tls-ca-bundle", "/etc/ssl/certs/openstack-ca.crt", "ca.crt"))
}
