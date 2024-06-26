package keystone

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	corev1 "k8s.io/api/core/v1"

	"github.com/ianunruh/openstack-operator/pkg/template"
)

func ClientEnv(prefix, secret string) []corev1.EnvVar {
	return []corev1.EnvVar{
		template.SecretEnvVar(prefix+"AUTH_URL", secret, "OS_AUTH_URL"),
		template.SecretEnvVar(prefix+"PASSWORD", secret, "OS_PASSWORD"),
		template.SecretEnvVar(prefix+"PROJECT_NAME", secret, "OS_PROJECT_NAME"),
		template.SecretEnvVar(prefix+"PROJECT_DOMAIN_NAME", secret, "OS_PROJECT_DOMAIN_NAME"),
		template.SecretEnvVar(prefix+"USER_DOMAIN_NAME", secret, "OS_USER_DOMAIN_NAME"),
		template.SecretEnvVar(prefix+"USERNAME", secret, "OS_USERNAME"),
	}
}

func MiddlewareEnv(prefix, secret string) []corev1.EnvVar {
	return append(ClientEnv(prefix, secret),
		template.SecretEnvVar(prefix+"WWW_AUTHENTICATE_URI", secret, "OS_AUTH_URL_WWW"))
}

func CloudClient(ctx context.Context, svcUser *corev1.Secret) (*gophercloud.ProviderClient, error) {
	authURL := svcUser.Data["OS_AUTH_URL"]
	if wwwAuthURL, ok := svcUser.Data["OS_AUTH_URL_WWW"]; ok {
		authURL = wwwAuthURL
	}

	client, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: string(authURL),
		Username:         string(svcUser.Data["OS_USERNAME"]),
		Password:         string(svcUser.Data["OS_PASSWORD"]),
		TenantName:       string(svcUser.Data["OS_PROJECT_NAME"]),
		DomainName:       string(svcUser.Data["OS_USER_DOMAIN_NAME"]),
	})
	if err != nil {
		return nil, fmt.Errorf("creating openstack client: %w", err)
	}

	client.Context = ctx

	return client, nil
}

func CloudEndpointOpts(svcUser *corev1.Secret) gophercloud.EndpointOpts {
	return gophercloud.EndpointOpts{
		Region:       string(svcUser.Data["OS_REGION_NAME"]),
		Availability: gophercloud.AvailabilityPublic,
	}
}

func NewIdentityServiceClient(ctx context.Context, svcUser *corev1.Secret) (*gophercloud.ServiceClient, error) {
	client, err := CloudClient(ctx, svcUser)
	if err != nil {
		return nil, err
	}

	return openstack.NewIdentityV3(client, CloudEndpointOpts(svcUser))
}
