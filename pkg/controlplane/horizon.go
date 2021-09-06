package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Horizon(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Horizon {
	spec := instance.Spec.Horizon

	spec.Server.Ingress = ingressDefaults(spec.Server.Ingress, instance, "horizon")
	spec.Server.NodeSelector = controllerNodeSelector(spec.Server.NodeSelector, instance)

	return &openstackv1beta1.Horizon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "horizon",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
