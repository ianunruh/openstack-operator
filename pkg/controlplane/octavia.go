package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Octavia(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Octavia {
	// TODO labels
	spec := instance.Spec.Octavia

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "octavia")

	return &openstackv1beta1.Octavia{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "octavia",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}