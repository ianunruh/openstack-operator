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
	ingress := ingressDefaults(instance.Spec.Keystone.API.Ingress, instance, "keystone")
	return fmt.Sprintf("https://%s/v3", ingress.Host)
}

func keystoneOIDCDashboardURL(instance *openstackv1beta1.ControlPlane) string {
	ingress := ingressDefaults(instance.Spec.Horizon.Server.Ingress, instance, "horizon")
	return fmt.Sprintf("https://%s/horizon/auth/websso/", ingress.Host)
}

func keystoneOIDCRedirectURI(instance *openstackv1beta1.ControlPlane) string {
	ingress := ingressDefaults(instance.Spec.Keystone.API.Ingress, instance, "keystone")
	return fmt.Sprintf("https://%s/v3/auth/OS-FEDERATION/websso/openid/redirect", ingress.Host)
}
