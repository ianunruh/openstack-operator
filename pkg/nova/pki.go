package nova

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.Nova) []*unstructured.Unstructured {
	return []*unstructured.Unstructured{
		CARootCertificate(instance),
		CAIssuer(instance),
		SelfSignedIssuer(instance),
	}
}

func ComputeCertificate(instance *openstackv1beta1.NovaComputeNode) *unstructured.Unstructured {
	return pki.Certificate(pki.CertificateParams{
		Name:       instance.Name,
		Namespace:  instance.Namespace,
		CommonName: instance.Spec.Node,
		SecretName: instance.Name,
		// TODO make configurable
		IssuerName: "nova-ca",
		Usages: []string{
			"digital signature",
			"key encipherment",
		},
	})
}

func CARootCertificate(instance *openstackv1beta1.Nova) *unstructured.Unstructured {
	return pki.Certificate(pki.CertificateParams{
		Name:       template.Combine(instance.Name, "ca-root"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(instance.Name, "ca-root"),
		IssuerName: template.Combine(instance.Name, "self-signed"),
		IsCA:       true,
	})
}

func CAIssuer(instance *openstackv1beta1.Nova) *unstructured.Unstructured {
	return pki.CAIssuer(pki.IssuerParams{
		Name:       template.Combine(instance.Name, "ca"),
		Namespace:  instance.Namespace,
		SecretName: template.Combine(instance.Name, "ca-root"),
	})
}

func SelfSignedIssuer(instance *openstackv1beta1.Nova) *unstructured.Unstructured {
	return pki.SelfSignedIssuer(pki.IssuerParams{
		Name:      template.Combine(instance.Name, "self-signed"),
		Namespace: instance.Namespace,
	})
}
