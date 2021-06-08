package rabbitmq

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "rabbitmq"
)

func ConfigMap(instance *openstackv1beta1.RabbitMQ) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["rabbitmq.conf"] = template.MustReadFile(AppLabel, "rabbitmq.conf")

	return cm
}

func Secret(instance *openstackv1beta1.RabbitMQ) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	password := template.NewPassword()

	secret.StringData["erlang-cookie"] = template.NewPassword()
	secret.StringData["password"] = password
	secret.StringData["connection"] = fmt.Sprintf("rabbit://admin:%s@%s:15672", password, instance.Name)

	return secret
}
