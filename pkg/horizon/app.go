package horizon

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "horizon"
)

func ConfigMap(instance *openstackv1beta1.Horizon) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["httpd.conf"] = template.MustReadFile(AppLabel, "httpd.conf")
	cm.Data["kolla.json"] = template.MustReadFile(AppLabel, "kolla.json")
	cm.Data["local_settings.py"] = template.MustRenderFile(AppLabel, "local_settings.py", configParamsFrom(instance))

	return cm
}

func Secret(instance *openstackv1beta1.Horizon) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	secret.StringData["secret-key"] = template.MustGeneratePassword()

	return secret
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Horizon, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Horizon) {
		instance.Spec = intended.Spec
	})
}

type configParams struct {
	SSO configSSOParams
}

type configSSOParams struct {
	Enabled       bool
	KeystoneURL   string
	InitialChoice string
	Choices       []configSSOChoice
}

type configSSOChoice struct {
	Kind  string
	Title string
}

func configParamsFrom(instance *openstackv1beta1.Horizon) configParams {
	params := configParams{}

	if ssoSpec := instance.Spec.SSO; ssoSpec.Enabled {
		var (
			initialChoice string
			choices       []configSSOChoice
		)

		for _, method := range ssoSpec.Methods {
			if method.Default {
				initialChoice = method.Kind
			}
			choices = append(choices, configSSOChoice{
				Kind:  method.Kind,
				Title: method.Title,
			})
		}

		params.SSO = configSSOParams{
			Enabled:       true,
			KeystoneURL:   ssoSpec.KeystoneURL,
			InitialChoice: initialChoice,
			Choices:       choices,
		}
	}

	return params
}
