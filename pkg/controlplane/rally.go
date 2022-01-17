package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Rally(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Rally {
	// TODO labels
	spec := instance.Spec.Rally

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	return &openstackv1beta1.Rally{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rally",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
