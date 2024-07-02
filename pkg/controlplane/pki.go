package controlplane

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.ControlPlane) []*unstructured.Unstructured {
	spec := instance.Spec.TLS
	if spec.Disabled || spec.Server.Issuer.Name == "" {
		return nil
	}

	return []*unstructured.Unstructured{
		CARootCertificate(instance),
		CAIssuer(instance),
		SelfSignedIssuer(instance),
	}
}

func CARootCertificate(instance *openstackv1beta1.ControlPlane) *unstructured.Unstructured {
	spec := instance.Spec.TLS.Server
	return pki.Certificate(pki.CertificateParams{
		Name:       template.Combine(spec.Issuer.Name, "root"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(spec.Issuer.Name, "root"),
		IssuerName: template.Combine(instance.Name, "self-signed"),
		IsCA:       true,
	})
}

func CAIssuer(instance *openstackv1beta1.ControlPlane) *unstructured.Unstructured {
	spec := instance.Spec.TLS.Server
	return pki.CAIssuer(pki.IssuerParams{
		Name:       spec.Issuer.Name,
		Namespace:  instance.Namespace,
		SecretName: template.Combine(spec.Issuer.Name, "root"),
	})
}

func SelfSignedIssuer(instance *openstackv1beta1.ControlPlane) *unstructured.Unstructured {
	return pki.SelfSignedIssuer(pki.IssuerParams{
		Name:      template.Combine(instance.Name, "self-signed"),
		Namespace: instance.Namespace,
	})
}

func tlsClientDefaults(spec openstackv1beta1.TLSClientSpec, instance *openstackv1beta1.ControlPlane) openstackv1beta1.TLSClientSpec {
	if instance.Spec.TLS.Disabled {
		return spec
	}

	if spec.CABundle == "" {
		spec.CABundle = instance.Spec.TLS.Client.CABundle
	}

	return spec
}

func tlsServerDefaults(spec openstackv1beta1.TLSServerSpec, instance *openstackv1beta1.ControlPlane) openstackv1beta1.TLSServerSpec {
	if instance.Spec.TLS.Disabled {
		return spec
	}

	if spec.Secret == "" {
		spec.Secret = instance.Spec.TLS.Server.Secret
	}

	if spec.Issuer.Name == "" {
		spec.Issuer.Name = instance.Spec.TLS.Server.Issuer.Name
	}
	if spec.Issuer.Kind == "" {
		spec.Issuer.Kind = instance.Spec.TLS.Server.Issuer.Kind
	}

	return spec
}
