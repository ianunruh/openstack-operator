package octavia

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	amphoraImageName         = "amphora"
	amphoraKeypairName       = "amphora"
	amphoraNetworkName       = "amphora"
	amphoraSecurityGroupName = "amphora"
)

func BootstrapAmphora(ctx context.Context, instance *openstackv1beta1.Octavia, c client.Client, log logr.Logger) error {
	adminSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "keystone"),
			Namespace: instance.Namespace,
		},
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(adminSecret), adminSecret); err != nil {
		return err
	}

	clientOpts := gophercloud.AuthOptions{
		IdentityEndpoint: string(adminSecret.Data["OS_AUTH_URL"]),
		Username:         string(adminSecret.Data["OS_USERNAME"]),
		Password:         string(adminSecret.Data["OS_PASSWORD"]),
		TenantName:       string(adminSecret.Data["OS_PROJECT_NAME"]),
		DomainName:       string(adminSecret.Data["OS_USER_DOMAIN_NAME"]),
	}

	client, err := openstack.AuthenticatedClient(clientOpts)
	if err != nil {
		return err
	}

	endpointOpts := gophercloud.EndpointOpts{
		Region: string(adminSecret.Data["OS_REGION_NAME"]),
	}

	compute, err := openstack.NewComputeV2(client, endpointOpts)
	if err != nil {
		return err
	}

	image, err := openstack.NewImageServiceV2(client, endpointOpts)
	if err != nil {
		return err
	}

	network, err := openstack.NewNetworkV2(client, endpointOpts)
	if err != nil {
		return err
	}

	b := &Bootstrap{
		client:   c,
		instance: instance,
		log:      log,

		compute: compute,
		image:   image,
		network: network,
	}

	return b.EnsureAll(ctx)
}

type Bootstrap struct {
	client   client.Client
	instance *openstackv1beta1.Octavia
	log      logr.Logger

	compute *gophercloud.ServiceClient
	image   *gophercloud.ServiceClient
	network *gophercloud.ServiceClient
}

func (b *Bootstrap) EnsureAll(ctx context.Context) error {
	if err := b.EnsureFlavor(ctx); err != nil {
		return err
	}

	if err := b.EnsureImage(ctx); err != nil {
		return err
	}

	if err := b.EnsureKeypair(ctx); err != nil {
		return err
	}

	if err := b.EnsureNetwork(ctx); err != nil {
		return err
	}

	if err := b.EnsureSecurityGroup(ctx); err != nil {
		return err
	}

	return nil
}

func (b *Bootstrap) EnsureFlavor(ctx context.Context) error {
	if b.instance.Status.Amphora.FlavorID != "" {
		return nil
	}

	// TODO make flavor opts configurable
	flavorDisk := 10

	flavor, err := flavors.Create(b.compute, flavors.CreateOpts{
		Name:  "c1-amphora",
		VCPUs: 2,
		RAM:   2048,
		Disk:  &flavorDisk,
	}).Extract()
	if err != nil {
		return err
	}

	b.instance.Status.Amphora.FlavorID = flavor.ID
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func (b *Bootstrap) EnsureImage(ctx context.Context) error {
	if b.instance.Status.Amphora.ImageProjectID != "" {
		return nil
	}

	// TODO handle image URL change
	// TODO create image
	// TODO update instance status

	return nil
}

func (b *Bootstrap) EnsureKeypair(ctx context.Context) error {
	result := keypairs.Get(b.compute, amphoraKeypairName, keypairs.GetOpts{})
	if err := result.Err; err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return err
		}
	} else {
		return nil
	}

	// TODO create keypair from secret
	// TODO update instance status

	return nil
}

func (b *Bootstrap) EnsureNetwork(ctx context.Context) error {
	if len(b.instance.Status.Amphora.NetworkIDs) > 0 {
		return nil
	}

	network, err := networks.Create(b.network, networks.CreateOpts{
		Name: amphoraNetworkName,
	}).Extract()
	if err != nil {
		return err
	}

	_, err = subnets.Create(b.network, subnets.CreateOpts{
		NetworkID: network.ID,
		CIDR:      "192.168.250.0/24",
	}).Extract()
	if err != nil {
		return err
	}

	b.instance.Status.Amphora.NetworkIDs = []string{network.ID}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func (b *Bootstrap) EnsureSecurityGroup(ctx context.Context) error {
	if len(b.instance.Status.Amphora.SecurityGroupIDs) > 0 {
		return nil
	}

	// TODO create security group
	// TODO update instance status

	return nil
}
