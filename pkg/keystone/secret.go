package keystone

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/gophercloud/utils/openstack/clientconfig"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func AdminSecret(instance *openstackv1beta1.Keystone, password string) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	ingress := instance.Spec.API.Ingress

	var authURL string
	if ingress == nil {
		authURL = fmt.Sprintf("http://%s-api.%s.svc:5000/v3", instance.Name, instance.Namespace)
	} else {
		authURL = fmt.Sprintf("https://%s/v3", ingress.Host)
	}

	if password == "" {
		password = template.MustGeneratePassword()
	}

	cloudsYAML := clientconfig.Clouds{
		Clouds: map[string]clientconfig.Cloud{
			"default": {
				AuthInfo: &clientconfig.AuthInfo{
					AuthURL:           authURL,
					Username:          "admin",
					Password:          password,
					ProjectName:       "admin",
					ProjectDomainName: "Default",
					UserDomainName:    "Default",
				},
				RegionName: "RegionOne",
			},
		},
	}

	secret.StringData = map[string]string{
		"OS_IDENTITY_API_VERSION": "3",
		"OS_AUTH_URL":             authURL,
		"OS_REGION_NAME":          "RegionOne",
		"OS_PROJECT_DOMAIN_NAME":  "Default",
		"OS_USER_DOMAIN_NAME":     "Default",
		"OS_PROJECT_NAME":         "admin",
		"OS_USERNAME":             "admin",
		"OS_PASSWORD":             password,
		"clouds.yaml":             string(template.MustEncodeYAML(cloudsYAML)),
	}

	return secret
}

func Secrets(instance *openstackv1beta1.Keystone) []*corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)

	return []*corev1.Secret{
		fernetSecret(template.Combine(instance.Name, "credential-keys"), instance.Namespace, labels),
		fernetSecret(template.Combine(instance.Name, "fernet-keys"), instance.Namespace, labels),
		memcacheSecret(template.Combine(instance.Name, "memcache"), instance.Namespace, labels),
	}
}

func fernetSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	secret := template.GenericSecret(name, namespace, labels)
	secret.StringData["0"] = template.MustGenerateFernetKey()
	secret.StringData["1"] = template.MustGenerateFernetKey()
	return secret
}

func memcacheSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	secret := template.GenericSecret(name, namespace, labels)
	secret.StringData["secret-key"] = template.MustGeneratePassword()
	return secret
}
