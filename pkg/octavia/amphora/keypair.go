package amphora

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func newKeypairSecret(instance *openstackv1beta1.Octavia) (*corev1.Secret, error) {
	labels := template.AppLabels(instance.Name, "octavia")
	name := template.Combine(instance.Name, "amphora-ssh")

	return template.SSHKeypairSecret(name, instance.Namespace, labels)
}
