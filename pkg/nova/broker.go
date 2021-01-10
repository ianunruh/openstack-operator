package nova

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func BrokerUser(name, namespace string, spec openstackv1beta1.RabbitMQUserSpec) *openstackv1beta1.RabbitMQUser {
	return &openstackv1beta1.RabbitMQUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}
