package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Manila(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Manila {
	// TODO labels
	spec := instance.Spec.Manila

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "manila")

	return &openstackv1beta1.Manila{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manila",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
