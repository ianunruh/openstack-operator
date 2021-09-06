package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Placement(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Placement {
	// TODO labels
	spec := instance.Spec.Placement

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "placement")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	return &openstackv1beta1.Placement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "placement",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
