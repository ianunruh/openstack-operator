package heat

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PKIResources(instance *openstackv1beta1.Heat) []*unstructured.Unstructured {
	var resources []*unstructured.Unstructured
	if cert := APICertificate(instance); cert != nil {
		resources = append(resources, cert)
	}
	if cert := CFNCertificate(instance); cert != nil {
		resources = append(resources, cert)
	}
	return resources
}

func APICertificate(instance *openstackv1beta1.Heat) *unstructured.Unstructured {
	name := template.Combine(instance.Name, "api")
	return pki.ServerCertificate(name, instance.Namespace, instance.Spec.API.TLS)
}

func CFNCertificate(instance *openstackv1beta1.Heat) *unstructured.Unstructured {
	name := template.Combine(instance.Name, "cfn")
	return pki.ServerCertificate(name, instance.Namespace, instance.Spec.CFN.TLS)
}
