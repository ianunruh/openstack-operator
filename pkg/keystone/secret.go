package keystone

import (
	"encoding/base64"
	"math/rand"
	"time"

	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Secrets(instance *openstackv1beta1.Keystone) []*corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)

	return []*corev1.Secret{
		adminSecret(instance.Name, instance.Namespace, labels),
		fernetSecret(template.Combine(instance.Name, "credential-keys"), instance.Namespace, labels),
		fernetSecret(template.Combine(instance.Name, "fernet-keys"), instance.Namespace, labels),
	}
}

func adminSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	secret := template.GenericSecret(name, namespace, labels)
	secret.StringData["password"] = template.NewPassword()
	return secret
}

func fernetSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	secret := template.GenericSecret(name, namespace, labels)
	secret.StringData["0"] = newFernetKey()
	secret.StringData["1"] = newFernetKey()
	return secret
}

func newFernetKey() string {
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, 32)
	for i := 0; i < 32; i++ {
		data[i] = byte(rand.Intn(10))
	}
	return base64.StdEncoding.EncodeToString(data)
}
