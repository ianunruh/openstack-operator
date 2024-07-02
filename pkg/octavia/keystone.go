package octavia

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneService(instance *openstackv1beta1.Octavia) *openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneServiceSpec{
			Name:        "octavia",
			Type:        "load-balancer",
			InternalURL: APIInternalURL(instance),
			PublicURL:   APIPublicURL(instance),
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Octavia) *openstackv1beta1.KeystoneUser {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneUserSpec{
			Secret:  template.Combine(instance.Name, "keystone"),
			Project: "service",
		},
	}
}

func EnsureKeystoneRoles(ctx context.Context, instance *openstackv1beta1.Octavia, c client.Client) error {
	adminUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(adminUser), adminUser); err != nil {
		return err
	}

	clientOpts := gophercloud.AuthOptions{
		IdentityEndpoint: string(adminUser.Data["OS_AUTH_URL"]),
		Username:         string(adminUser.Data["OS_USERNAME"]),
		Password:         string(adminUser.Data["OS_PASSWORD"]),
		TenantName:       string(adminUser.Data["OS_PROJECT_NAME"]),
		DomainName:       string(adminUser.Data["OS_USER_DOMAIN_NAME"]),
	}

	client, err := openstack.AuthenticatedClient(clientOpts)
	if err != nil {
		return err
	}

	endpointOpts := gophercloud.EndpointOpts{
		Region:       string(adminUser.Data["OS_REGION_NAME"]),
		Availability: gophercloud.AvailabilityPublic,
	}

	identity, err := openstack.NewIdentityV3(client, endpointOpts)
	if err != nil {
		return err
	}

	allPages, err := roles.List(identity, roles.ListOpts{}).AllPages()
	if err != nil {
		return err
	}

	allRoles, err := roles.ExtractRoles(allPages)
	if err != nil {
		return err
	}

	currentRoleNames := make(map[string]bool, len(allRoles))
	for _, role := range allRoles {
		currentRoleNames[role.Name] = true
	}

	roleNames := []string{
		"load-balancer_admin",
		"load-balancer_observer",
		"load-balancer_global_observer",
		"load-balancer_quota_admin",
		"load-balancer_member",
	}
	for _, name := range roleNames {
		if currentRoleNames[name] {
			continue
		}

		_, err := roles.Create(identity, roles.CreateOpts{
			Name: name,
		}).Extract()
		if err != nil {
			return err
		}
	}

	// TODO Assign load-balancer_admin role to admin user

	return nil
}
