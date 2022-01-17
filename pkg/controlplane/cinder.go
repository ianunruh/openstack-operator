package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Cinder(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Cinder {
	// TODO labels
	spec := instance.Spec.Cinder

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "cinder")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	spec.Scheduler.NodeSelector = controllerNodeSelector(spec.Scheduler.NodeSelector, instance)

	spec.Volume.NodeSelector = controllerNodeSelector(spec.Volume.NodeSelector, instance)

	return &openstackv1beta1.Cinder{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cinder",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
