package nova

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func APIDatabase(instance *openstackv1beta1.Nova) *openstackv1beta1.MariaDBDatabase {
	return &openstackv1beta1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "api"),
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.APIDatabase,
	}
}

func CellDatabase(instanceName, cellName, namespace string, spec openstackv1beta1.MariaDBDatabaseSpec) *openstackv1beta1.MariaDBDatabase {
	return &openstackv1beta1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instanceName, cellName),
			Namespace: namespace,
		},
		Spec: spec,
	}
}
