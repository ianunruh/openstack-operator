package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Keystone(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Keystone {
	// TODO labels
	spec := instance.Spec.Keystone

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "keystone")

	return &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
