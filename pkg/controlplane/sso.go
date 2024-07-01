package controlplane

import (
	"fmt"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func horizonSSOMethods() []openstackv1beta1.HorizonSSOMethod {
	return []openstackv1beta1.HorizonSSOMethod{
		{
			Kind:    "credentials",
			Title:   "Keystone Credentials",
			Default: true,
		},
		{
			Kind:  "openid",
			Title: "OpenID Connect",
		},
	}
}

func horizonSSOKeystoneURL(instance *openstackv1beta1.ControlPlane) string {
	spec := instance.Spec.Keystone
	ingress := ingressDefaults(spec.API.Ingress, instance, "keystone")
	return fmt.Sprintf("https://%s/v3", ingress.Host)
}

func keystoneOIDCDashboardURL(instance *openstackv1beta1.ControlPlane) string {
	spec := instance.Spec.Horizon
	ingress := ingressDefaults(spec.Server.Ingress, instance, "horizon")
	return fmt.Sprintf("https://%s/auth/websso/", ingress.Host)
}

func keystoneOIDCRedirectURI(instance *openstackv1beta1.ControlPlane) string {
	spec := instance.Spec.Keystone
	ingress := ingressDefaults(spec.API.Ingress, instance, "keystone")
	return fmt.Sprintf(
		"https://%s/v3/OS-FEDERATION/identity_providers/%s/protocols/openid/auth",
		ingress.Host,
		spec.OIDC.IdentityProvider)
}
