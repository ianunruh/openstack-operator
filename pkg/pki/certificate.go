package pki

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

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
