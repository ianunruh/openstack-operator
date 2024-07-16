package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Broker(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.RabbitMQ {
	if instance.Spec.ExternalBroker != nil {
		return nil
	}

	spec := instance.Spec.Broker

	spec.Management.Ingress = ingressDefaults(spec.Management.Ingress, instance, "rabbitmq")

	spec.NodeSelector = controllerNodeSelector(spec.NodeSelector, instance)

	spec.TLS = tlsServerDefaults(spec.TLS, instance)

	// TODO labels
	return &openstackv1beta1.RabbitMQ{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

func brokerUserDefaults(spec openstackv1beta1.RabbitMQUserSpec, instance *openstackv1beta1.ControlPlane) openstackv1beta1.RabbitMQUserSpec {
	if spec.External == nil {
		spec.External = instance.Spec.ExternalBroker
	}
	return spec
}
