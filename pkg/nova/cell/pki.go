package cell

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.NovaCell) []*unstructured.Unstructured {
	var resources []*unstructured.Unstructured
	if cert := MetadataCertificate(instance); cert != nil {
		resources = append(resources, cert)
	}
	return resources
}

func MetadataCertificate(instance *openstackv1beta1.NovaCell) *unstructured.Unstructured {
	name := template.Combine(instance.Name, "metadata")
	return pki.ServerCertificate(name, instance.Namespace, instance.Spec.Metadata.TLS)
}
