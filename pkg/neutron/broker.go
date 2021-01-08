package neutron

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func BrokerUser(instance *openstackv1beta1.Neutron) *openstackv1beta1.RabbitMQUser {
	return &openstackv1beta1.RabbitMQUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.Broker,
	}
}
