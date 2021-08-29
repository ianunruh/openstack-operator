package amphora

import (
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
	corev1 "k8s.io/api/core/v1"
)

func Secret(instance *openstackv1beta1.Octavia) *corev1.Secret {
	name := template.Combine(instance.Name, "amphora")
	labels := template.AppLabels(instance.Name, "octavia")
	secret := template.GenericSecret(name, instance.Namespace, labels)

	secret.StringData["heartbeat-key"] = template.NewPassword()

	return secret
}
