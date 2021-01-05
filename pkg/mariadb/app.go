package mariadb

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "mariadb"
)

func ConfigMap(instance *openstackv1beta1.MariaDB) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["my.cnf"] = template.MustRenderFile(AppLabel, "my.cnf", nil)

	return cm
}

func Secret(instance *openstackv1beta1.MariaDB) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	secret.StringData["password"] = template.NewPassword()

	return secret
}
