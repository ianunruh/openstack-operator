package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Barbican(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Barbican {
	spec := instance.Spec.Barbican
	if spec == nil {
		return nil
	}

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "barbican")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)
	spec.API.TLS = tlsServerDefaults(spec.API.TLS, instance)

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	spec.Worker.NodeSelector = controllerNodeSelector(spec.Worker.NodeSelector, instance)

	spec.Broker = brokerUserDefaults(spec.Broker, instance)
	spec.Cache = cacheDefaults(spec.Cache, instance)
	spec.Database = databaseDefaults(spec.Database, instance)
	spec.TLS = tlsClientDefaults(spec.TLS, instance)

	// TODO labels
	return &openstackv1beta1.Barbican{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "barbican",
			Namespace: instance.Namespace,
		},
		Spec: *spec,
	}
}
