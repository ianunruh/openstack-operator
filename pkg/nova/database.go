package nova

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Database(instance *openstackv1beta1.Nova) *openstackv1beta1.MariaDBDatabase {
	return &openstackv1beta1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.Database,
	}
}

func APIDatabase(instance *openstackv1beta1.Nova) *openstackv1beta1.MariaDBDatabase {
	return &openstackv1beta1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "api"),
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.APIDatabase,
	}
}

func CellDatabase(instance *openstackv1beta1.Nova) *openstackv1beta1.MariaDBDatabase {
	return &openstackv1beta1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "cell0"),
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.CellDatabase,
	}
}
