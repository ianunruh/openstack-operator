package keystone

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "keystone"
)

var (
	appUID = int64(42425)
)

func ConfigMap(instance *openstackv1beta1.Keystone) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "keystone.conf")

	cfg.Section("cache").NewKey("backend_argument",
		fmt.Sprintf("url:%s", strings.Join(spec.Cache.Servers, ",")))

	if spec.Notifications.Enabled {
		cfg.Section("oslo_messaging_notifications").NewKey("driver", "messagingv2")
	}

	if spec.OIDC.Enabled {
		cfg.Section("auth").NewKey("methods", "password,token,oauth1,openid,mapped,application_credential")
		cfg.Section("federation").NewKey("trusted_dashboard", spec.OIDC.DashboardURL)
	}

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["httpd.conf"] = template.MustRenderFile(AppLabel, "httpd.conf", httpdParamsFrom(instance))
	cm.Data["kolla.json"] = template.MustReadFile(AppLabel, "kolla.json")

	cm.Data["keystone.conf"] = template.MustOutputINI(cfg).String()

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Keystone, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Keystone) {
		instance.Spec = intended.Spec
	})
}

type httpdParams struct {
	TLS bool

	OIDC httpdOIDCParams
}

type httpdOIDCParams struct {
	Enabled             bool
	ExtraConfig         map[string]string
	IdentityProvider    string
	ProviderMetadataURL string
	RedirectURI         string
	RequireClaims       []string
	Scopes              string
}

func httpdParamsFrom(instance *openstackv1beta1.Keystone) httpdParams {
	params := httpdParams{
		TLS: instance.Spec.API.TLS.Secret != "",
	}

	if oidcSpec := instance.Spec.OIDC; oidcSpec.Enabled {
		params.OIDC = httpdOIDCParams{
			Enabled:             true,
			ExtraConfig:         oidcSpec.ExtraConfig,
			IdentityProvider:    oidcSpec.IdentityProvider,
			ProviderMetadataURL: oidcSpec.ProviderMetadataURL,
			RedirectURI:         oidcSpec.RedirectURI,
			RequireClaims:       oidcSpec.RequireClaims,
			Scopes:              strings.Join(oidcSpec.Scopes, " "),
		}
	}

	return params
}
