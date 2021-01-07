package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Broker(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.RabbitMQ {
	// TODO labels
	return &openstackv1beta1.RabbitMQ{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.Broker,
	}
}
