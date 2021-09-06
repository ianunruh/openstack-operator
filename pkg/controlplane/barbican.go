package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Barbican(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Barbican {
	// TODO labels
	spec := instance.Spec.Barbican

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "barbican")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	spec.Worker.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	return &openstackv1beta1.Barbican{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "barbican",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
