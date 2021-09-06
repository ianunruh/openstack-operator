package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func OVNControlPlane(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.OVNControlPlane {
	// TODO labels
	spec := instance.Spec.OVN

	spec.Northd.NodeSelector = controllerNodeSelector(spec.Northd.NodeSelector, instance)
	spec.OVSDBNorth.NodeSelector = controllerNodeSelector(spec.OVSDBNorth.NodeSelector, instance)
	spec.OVSDBSouth.NodeSelector = controllerNodeSelector(spec.OVSDBSouth.NodeSelector, instance)
	spec.Node.NodeSelector = computeNodeSelector(spec.Node.NodeSelector, instance)

	return &openstackv1beta1.OVNControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ovn",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
