package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Magnum(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Magnum {
	spec := instance.Spec.Magnum
	if spec == nil {
		return nil
	}

	spec.Image = imageDefault(spec.Image, DefaultMagnumImage)

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "magnum")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	spec.Conductor.NodeSelector = controllerNodeSelector(spec.Conductor.NodeSelector, instance)

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	// TODO labels
	return &openstackv1beta1.Magnum{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "magnum",
			Namespace: instance.Namespace,
		},
		Spec: *spec,
	}
}
