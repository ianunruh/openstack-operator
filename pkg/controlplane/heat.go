package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Heat(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Heat {
	// TODO labels
	spec := instance.Spec.Heat

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "heat")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	spec.CFN.Ingress = ingressDefaults(spec.CFN.Ingress, instance, "heat-cfn")
	spec.CFN.NodeSelector = controllerNodeSelector(spec.CFN.NodeSelector, instance)

	return &openstackv1beta1.Heat{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "heat",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
