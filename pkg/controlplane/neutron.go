package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Neutron(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Neutron {
	// TODO labels
	spec := instance.Spec.Neutron

	spec.Image = imageDefault(spec.Image, DefaultNeutronImage)

	spec.Server.Ingress = ingressDefaults(spec.Server.Ingress, instance, "neutron")
	spec.Server.NodeSelector = controllerNodeSelector(spec.Server.NodeSelector, instance)

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	spec.MetadataAgent.NodeSelector = computeNodeSelector(spec.MetadataAgent.NodeSelector, instance)

	return &openstackv1beta1.Neutron{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "neutron",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
