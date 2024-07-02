package horizon

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.Horizon) []*unstructured.Unstructured {
	var resources []*unstructured.Unstructured
	if cert := ServerCertificate(instance); cert != nil {
		resources = append(resources, cert)
	}
	return resources
}

func ServerCertificate(instance *openstackv1beta1.Horizon) *unstructured.Unstructured {
	name := template.Combine(instance.Name, "server")
	return pki.ServerCertificate(name, instance.Name, instance.Spec.Server.TLS)
}
