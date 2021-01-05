package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Database(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.MariaDB {
	// TODO labels
	return &openstackv1beta1.MariaDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mariadb",
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.Database,
	}
}
