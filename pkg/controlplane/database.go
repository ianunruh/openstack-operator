package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Database(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.MariaDB {
	if instance.Spec.ExternalDatabase != nil {
		return nil
	}

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

func databaseDefaults(spec openstackv1beta1.MariaDBDatabaseSpec, instance *openstackv1beta1.ControlPlane) openstackv1beta1.MariaDBDatabaseSpec {
	if spec.External == nil {
		spec.External = instance.Spec.ExternalDatabase
	}
	return spec
}
