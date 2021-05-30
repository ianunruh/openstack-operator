package keystone

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Secrets(instance *openstackv1beta1.Keystone) []*corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)

	return []*corev1.Secret{
		adminSecret(instance.Name, instance.Namespace, labels, instance.Spec.API.Ingress),
		fernetSecret(template.Combine(instance.Name, "credential-keys"), instance.Namespace, labels),
		fernetSecret(template.Combine(instance.Name, "fernet-keys"), instance.Namespace, labels),
	}
}

func adminSecret(name, namespace string, labels map[string]string, ingress *openstackv1beta1.IngressSpec) *corev1.Secret {
	secret := template.GenericSecret(name, namespace, labels)

	var authURL string
	if ingress == nil {
		authURL = fmt.Sprintf("http://%s-api.%s.svc:5000/v3", name, namespace)
	} else {
		authURL = fmt.Sprintf("https://%s/v3", ingress.Host)
	}

	secret.StringData = map[string]string{
		"OS_IDENTITY_API_VERSION": "3",
		"OS_AUTH_URL":             authURL,
		"OS_REGION_NAME":          "RegionOne",
		"OS_PROJECT_DOMAIN_NAME":  "Default",
		"OS_USER_DOMAIN_NAME":     "Default",
		"OS_PROJECT_NAME":         "admin",
		"OS_USERNAME":             "admin",
		"OS_PASSWORD":             template.NewPassword(),
	}

	return secret
}

func fernetSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	secret := template.GenericSecret(name, namespace, labels)
	secret.StringData["0"] = template.NewFernetKey()
	secret.StringData["1"] = template.NewFernetKey()
	return secret
}
