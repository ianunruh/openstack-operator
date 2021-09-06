package controlplane

import (
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func controllerNodeSelector(selector map[string]string, instance *openstackv1beta1.ControlPlane) map[string]string {
	if selector == nil {
		return instance.Spec.NodeSelector.Controller
	}

	return selector
}

func computeNodeSelector(selector map[string]string, instance *openstackv1beta1.ControlPlane) map[string]string {
	if selector == nil {
		return instance.Spec.NodeSelector.Compute
	}

	return selector
}
