package pki

import (
	"github.com/ianunruh/openstack-operator/pkg/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type IssuerParams struct {
	Name       string
	Namespace  string
	SecretName string
}

func CAIssuer(params IssuerParams) *unstructured.Unstructured {
	manifest := template.MustRenderFile("pki", "issuer-ca.yaml", params)
	return template.MustDecodeManifest(manifest)
}

func SelfSignedIssuer(params IssuerParams) *unstructured.Unstructured {
	manifest := template.MustRenderFile("pki", "issuer-self-signed.yaml", params)
	return template.MustDecodeManifest(manifest)
}
