package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Magnum(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Magnum {
	// TODO labels
	spec := instance.Spec.Magnum

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "magnum")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	spec.Conductor.NodeSelector = controllerNodeSelector(spec.Conductor.NodeSelector, instance)

	return &openstackv1beta1.Magnum{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "magnum",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
