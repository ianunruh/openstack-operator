package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Cache(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Memcached {
	// TODO labels
	return &openstackv1beta1.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "memcached",
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.Cache,
	}
}
