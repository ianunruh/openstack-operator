package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Broker(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.RabbitMQ {
	spec := instance.Spec.Broker

	spec.Management.Ingress = ingressDefaults(spec.Management.Ingress, instance, "rabbitmq")

	spec.NodeSelector = controllerNodeSelector(spec.NodeSelector, instance)

	// TODO labels
	return &openstackv1beta1.RabbitMQ{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
