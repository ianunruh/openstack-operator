package rabbitmq

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
)

func PKIResources(instance *openstackv1beta1.RabbitMQ) []*unstructured.Unstructured {
	var resources []*unstructured.Unstructured
	if cert := ServerCertificate(instance); cert != nil {
		resources = append(resources, cert)
	}
	return resources
}

func ServerCertificate(instance *openstackv1beta1.RabbitMQ) *unstructured.Unstructured {
	return pki.ServerCertificate(instance.Name, instance.Namespace, instance.Spec.TLS)
}
