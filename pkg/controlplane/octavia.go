package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Octavia(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Octavia {
	spec := instance.Spec.Octavia
	if spec == nil {
		return nil
	}

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "octavia")
	spec.API.TLS = tlsServerDefaults(spec.API.TLS, instance)

	spec.NodeSelector = controllerNodeSelector(spec.NodeSelector, instance)

	spec.Broker = brokerUserDefaults(spec.Broker, instance)
	spec.Cache = cacheDefaults(spec.Cache, instance)
	spec.Database = databaseDefaults(spec.Database, instance)
	spec.TLS = tlsClientDefaults(spec.TLS, instance)

	// TODO labels
	return &openstackv1beta1.Octavia{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "octavia",
			Namespace: instance.Namespace,
		},
		Spec: *spec,
	}
}
