package ovn

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.OVNControlPlane) []*unstructured.Unstructured {
	return []*unstructured.Unstructured{
		CARootCertificate(instance),
		CAIssuer(instance),
		SelfSignedIssuer(instance),
	}
}

type certificateParams struct {
	Name       string
	SecretName string
	IssuerName string
}

func CARootCertificate(instance *openstackv1beta1.OVNControlPlane) *unstructured.Unstructured {
	manifest := template.MustRenderFile(AppLabel, "certificate-ca-root.yaml", certificateParams{
		Name:       template.Combine(instance.Name, "ca-root"),
		SecretName: template.Combine(instance.Name, "ca-root"),
		IssuerName: template.Combine(instance.Name, "self-signed"),
	})

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(instance.Namespace)
	return res
}

type issuerParams struct {
	Name       string
	SecretName string
}

func CAIssuer(instance *openstackv1beta1.OVNControlPlane) *unstructured.Unstructured {
	manifest := template.MustRenderFile(AppLabel, "issuer-ca.yaml", issuerParams{
		Name:       template.Combine(instance.Name, "ca"),
		SecretName: template.Combine(instance.Name, "ca-root"),
	})

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(instance.Namespace)
	return res
}

func SelfSignedIssuer(instance *openstackv1beta1.OVNControlPlane) *unstructured.Unstructured {
	manifest := template.MustRenderFile(AppLabel, "issuer-self-signed.yaml", issuerParams{
		Name: template.Combine(instance.Name, "self-signed"),
	})

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(instance.Namespace)
	return res
}
