package keystone

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.Keystone) []*unstructured.Unstructured {
	if instance.Spec.API.TLS.Secret == "" {
		return nil
	}

	return []*unstructured.Unstructured{
		APICertificate(instance),
	}
}

func APICertificate(instance *openstackv1beta1.Keystone) *unstructured.Unstructured {
	name := template.Combine(instance.Name, "api")
	spec := instance.Spec.API
	return pki.Certificate(pki.CertificateParams{
		Name:       name,
		Namespace:  instance.Namespace,
		SecretName: spec.TLS.Secret,
		IssuerName: spec.TLS.Issuer.Name,
		IssuerKind: spec.TLS.Issuer.Kind,
		DNSNames: []string{
			fmt.Sprintf("%s.%s.svc", name, instance.Namespace),
		},
	})
}
