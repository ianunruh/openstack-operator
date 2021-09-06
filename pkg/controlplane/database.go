package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Database(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.MariaDB {
	spec := instance.Spec.Database

	spec.NodeSelector = controllerNodeSelector(spec.NodeSelector, instance)

	// TODO labels
	return &openstackv1beta1.MariaDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mariadb",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}
