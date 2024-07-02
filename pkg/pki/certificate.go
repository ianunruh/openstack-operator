package pki

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

type CertificateParams struct {
	Name       string
	Namespace  string
	SecretName string
	IssuerName string
	IssuerKind string
	DNSNames   []string
	IsCA       bool
	Usages     []string
}

func Certificate(params CertificateParams) *unstructured.Unstructured {
	if params.IssuerKind == "" {
		params.IssuerKind = "Issuer"
	}

	manifest := template.MustRenderFile("pki", "certificate.yaml", params)
	return template.MustDecodeManifest(manifest)
}

func ServerCertificate(name, namespace string, spec openstackv1beta1.TLSServerSpec) *unstructured.Unstructured {
	if spec.Secret == "" || spec.Issuer.Name == "" {
		return nil
	}

	return Certificate(CertificateParams{
		Name:       name,
		Namespace:  namespace,
		SecretName: spec.Secret,
		IssuerName: spec.Issuer.Name,
		IssuerKind: spec.Issuer.Kind,
		DNSNames: []string{
			fmt.Sprintf("%s.%s.svc", name, namespace),
		},
	})
}
