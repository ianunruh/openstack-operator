package amphora

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.Octavia) []*unstructured.Unstructured {
	return []*unstructured.Unstructured{
		ClientCARootCertificate(instance),
		ServerCARootCertificate(instance),
		ClientCAIssuer(instance),
		ClientCertificate(instance),
		SelfSignedIssuer(instance),
	}
}

func ClientCertificate(instance *openstackv1beta1.Octavia) *unstructured.Unstructured {
	return pki.Certificate(pki.CertificateParams{
		Name:       template.Combine(instance.Name, "amphora-client"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(instance.Name, "amphora-client"),
		IssuerName: template.Combine(instance.Name, "amphora-client-ca"),
		Usages: []string{
			// https://docs.openstack.org/octavia/latest/admin/guides/certificates.html
			"client auth",
			"digital signature",
			"email protection",
			"key encipherment",
		},
	})
}

func ClientCARootCertificate(instance *openstackv1beta1.Octavia) *unstructured.Unstructured {
	return pki.Certificate(pki.CertificateParams{
		Name:       template.Combine(instance.Name, "amphora-client-ca-root"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(instance.Name, "amphora-client-ca-root"),
		IssuerName: template.Combine(instance.Name, "amphora-self-signed"),
		IsCA:       true,
	})
}

func ServerCARootCertificate(instance *openstackv1beta1.Octavia) *unstructured.Unstructured {
	return pki.Certificate(pki.CertificateParams{
		Name:       template.Combine(instance.Name, "amphora-server-ca-root"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(instance.Name, "amphora-server-ca-root"),
		IssuerName: template.Combine(instance.Name, "amphora-self-signed"),
		IsCA:       true,
	})
}

func ClientCAIssuer(instance *openstackv1beta1.Octavia) *unstructured.Unstructured {
	return pki.CAIssuer(pki.IssuerParams{
		Name:       template.Combine(instance.Name, "amphora-client-ca"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(instance.Name, "amphora-client-ca-root"),
	})
}

func SelfSignedIssuer(instance *openstackv1beta1.Octavia) *unstructured.Unstructured {
	return pki.SelfSignedIssuer(pki.IssuerParams{
		Name:      template.Combine(instance.Name, "amphora-self-signed"),
		Namespace: instance.Namespace,
	})
}
