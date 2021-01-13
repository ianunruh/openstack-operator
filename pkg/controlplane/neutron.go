package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Neutron(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Neutron {
	// TODO labels
	spec := instance.Spec.Neutron

	spec.Server.Ingress = ingressDefaults(spec.Server.Ingress, instance, "neutron")

	return &openstackv1beta1.Neutron{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "neutron",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
