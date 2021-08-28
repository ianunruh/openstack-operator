package manila

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Database(instance *openstackv1beta1.Manila) *openstackv1beta1.MariaDBDatabase {
	return &openstackv1beta1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.Database,
	}
}
