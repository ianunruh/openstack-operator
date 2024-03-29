package amphora

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	novakeypair "github.com/ianunruh/openstack-operator/pkg/nova/keypair"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	amphoraImageName   = "amphora"
	amphoraImageTag    = "amphora"
	amphoraKeypairName = "amphora"
	amphoraNetworkName = "octavia-lb-mgmt"

	imageSourceProperty = "source"
)

func Bootstrap(ctx context.Context, instance *openstackv1beta1.Octavia, c client.Client, log logr.Logger) (ctrl.Result, error) {
	b, err := newBootstrap(ctx, instance, c, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := b.EnsureAll(ctx); err != nil {
		return ctrl.Result{}, err
	}

	result := b.Wait()
	return result, nil
}

func newBootstrap(ctx context.Context, instance *openstackv1beta1.Octavia, c client.Client, log logr.Logger) (*bootstrap, error) {
	adminUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(adminUser), adminUser); err != nil {
		return nil, err
	}

	svcUser := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, "keystone"),
			Namespace: instance.Namespace,
		},
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(svcUser), svcUser); err != nil {
		return nil, err
	}

	clientOpts := gophercloud.AuthOptions{
		IdentityEndpoint: string(adminUser.Data["OS_AUTH_URL"]),
		Username:         string(svcUser.Data["OS_USERNAME"]),
		Password:         string(svcUser.Data["OS_PASSWORD"]),
		TenantName:       string(svcUser.Data["OS_PROJECT_NAME"]),
		DomainName:       string(svcUser.Data["OS_USER_DOMAIN_NAME"]),
	}

	client, err := openstack.AuthenticatedClient(clientOpts)
	if err != nil {
		return nil, err
	}

	endpointOpts := gophercloud.EndpointOpts{
		Region:       string(svcUser.Data["OS_REGION_NAME"]),
		Availability: gophercloud.AvailabilityPublic,
	}

	compute, err := openstack.NewComputeV2(client, endpointOpts)
	if err != nil {
		return nil, err
	}

	image, err := openstack.NewImageServiceV2(client, endpointOpts)
	if err != nil {
		return nil, err
	}

	network, err := openstack.NewNetworkV2(client, endpointOpts)
	if err != nil {
		return nil, err
	}

	b := &bootstrap{
		client:   c,
		deps:     template.NewConditionWaiter(log),
		instance: instance,
		log:      log,

		compute: compute,
		image:   image,
		network: network,
	}

	return b, nil
}

type bootstrap struct {
	client   client.Client
	deps     *template.ConditionWaiter
	instance *openstackv1beta1.Octavia
	log      logr.Logger

	compute *gophercloud.ServiceClient
	image   *gophercloud.ServiceClient
	network *gophercloud.ServiceClient
}

func (b *bootstrap) EnsureAll(ctx context.Context) error {
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

	if err := b.EnsureHealthSecurityGroup(ctx); err != nil {
		return err
	}

	if err := b.EnsureHealthPort(ctx); err != nil {
		return err
	}

	return nil
}

func (b *bootstrap) Wait() ctrl.Result {
	return b.deps.Wait()
}

func (b *bootstrap) EnsureFlavor(ctx context.Context) error {
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

func (b *bootstrap) EnsureImage(ctx context.Context) error {
	imageURL := b.instance.Spec.Amphora.ImageURL

	image, err := b.getCurrentImage(ctx, imageURL)
	if err != nil {
		return err
	} else if image == nil {
		image, err = b.uploadImage(ctx, imageURL)
		if err != nil {
			return err
		}
	}

	b.instance.Status.Amphora.ImageProjectID = image.Owner
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}
	return nil
}

func (b *bootstrap) getCurrentImage(ctx context.Context, imageURL string) (*images.Image, error) {
	pager := images.List(b.image, images.ListOpts{
		Tags: []string{amphoraImageTag},
	})

	page, err := pager.AllPages()
	if err != nil {
		return nil, err
	}

	images, err := images.ExtractImages(page)
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		if image.Properties[imageSourceProperty] == imageURL {
			return &image, nil
		}
	}

	return nil, nil
}

func (b *bootstrap) uploadImage(ctx context.Context, imageURL string) (*images.Image, error) {
	b.log.Info("Creating image", "name", amphoraImageName)
	image, err := images.Create(b.image, images.CreateOpts{
		Name:            amphoraImageName,
		Tags:            []string{amphoraImageTag},
		ContainerFormat: "bare",
		DiskFormat:      "qcow2",
		Properties: map[string]string{
			imageSourceProperty: imageURL,
		},
	}).Extract()
	if err != nil {
		return nil, err
	}

	b.log.Info("Uploading image",
		"name", image.Name,
		"url", imageURL)
	data, err := fetchImage(imageURL)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	if err := imagedata.Upload(b.image, image.ID, data).Err; err != nil {
		return nil, err
	}

	return image, nil
}

func (b *bootstrap) EnsureKeypair(ctx context.Context) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(b.instance.Name, "amphora-ssh"),
			Namespace: b.instance.Namespace,
		},
	}
	if err := b.client.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		secret, err = newKeypairSecret(b.instance)
		if err != nil {
			return err
		}
		b.log.Info("Creating keypair secret", "name", secret.Name)
		if err := b.client.Create(ctx, secret); err != nil {
			return err
		}
	}

	keypair := &openstackv1beta1.NovaKeypair{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(b.instance.Name, "amphora"),
			Namespace: b.instance.Namespace,
		},
		Spec: openstackv1beta1.NovaKeypairSpec{
			Name:      amphoraKeypairName,
			PublicKey: string(secret.Data["id_rsa.pub"]),
			User:      b.instance.Name,
		},
	}
	if err := novakeypair.Ensure(ctx, b.client, keypair, b.log); err != nil {
		return err
	}

	novakeypair.AddReadyCheck(b.deps, keypair)

	return nil
}

func (b *bootstrap) EnsureNetwork(ctx context.Context) error {
	if len(b.instance.Status.Amphora.NetworkIDs) > 0 {
		// TODO check if current exists, otherwise recreate
		return nil
	}

	b.log.Info("Creating network", "name", amphoraNetworkName)
	network, err := networks.Create(b.network, networks.CreateOpts{
		Name: amphoraNetworkName,
	}).Extract()
	if err != nil {
		return err
	}

	b.log.Info("Creating subnet", "networkID", network.ID)
	_, err = subnets.Create(b.network, subnets.CreateOpts{
		NetworkID: network.ID,
		IPVersion: gophercloud.IPv4,
		CIDR:      b.instance.Spec.Amphora.ManagementCIDR,
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

func (b *bootstrap) EnsureHealthPort(ctx context.Context) error {
	status := b.instance.Status.Amphora

	if len(status.HealthPorts) > 0 {
		_, err := ports.Get(b.network, status.HealthPorts[0].ID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); !ok {
				return err
			}
		} else {
			return nil
		}
	}

	networkID := status.NetworkIDs[0]
	securityGroups := status.HealthSecurityGroupIDs

	b.log.Info("Creating port",
		"name", "octavia-health-manager",
		"networkID", networkID)
	port, err := ports.Create(b.network, ports.CreateOpts{
		Name:           "octavia-health-manager",
		NetworkID:      networkID,
		SecurityGroups: &securityGroups,
		DeviceOwner:    "Octavia:health-mgr",
	}).Extract()
	if err != nil {
		return err
	}

	b.instance.Status.Amphora.HealthPorts = []openstackv1beta1.OctaviaAmphoraHealthPort{
		{
			ID:         port.ID,
			MACAddress: port.MACAddress,
			IPAddress:  port.FixedIPs[0].IPAddress,
		},
	}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func (b *bootstrap) EnsureSecurityGroup(ctx context.Context) error {
	if len(b.instance.Status.Amphora.SecurityGroupIDs) > 0 {
		return nil
	}

	b.log.Info("Creating security group", "name", "octavia-lb-mgmt")
	group, err := groups.Create(b.network, groups.CreateOpts{
		Name: "octavia-lb-mgmt",
	}).Extract()
	if err != nil {
		return err
	}

	// TODO reconcile each rule independently
	ruleOpts := []rules.CreateOpts{
		{Protocol: rules.ProtocolICMP},
		{Protocol: rules.ProtocolTCP, PortRangeMin: 22, PortRangeMax: 22},
		{Protocol: rules.ProtocolTCP, PortRangeMin: 9443, PortRangeMax: 9443},
	}
	for _, opts := range ruleOpts {
		opts.SecGroupID = group.ID
		opts.Direction = rules.DirIngress
		opts.EtherType = rules.EtherType4
		if err := rules.Create(b.network, opts).Err; err != nil {
			return err
		}
	}

	b.instance.Status.Amphora.SecurityGroupIDs = []string{group.ID}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func (b *bootstrap) EnsureHealthSecurityGroup(ctx context.Context) error {
	if len(b.instance.Status.Amphora.HealthSecurityGroupIDs) > 0 {
		return nil
	}

	b.log.Info("Creating security group", "name", "octavia-lb-health-manager")
	group, err := groups.Create(b.network, groups.CreateOpts{
		Name: "octavia-lb-health-manager",
	}).Extract()
	if err != nil {
		return err
	}

	// TODO reconcile each rule independently
	ruleOpts := []rules.CreateOpts{
		{Protocol: rules.ProtocolUDP, PortRangeMin: 5555, PortRangeMax: 5555},
	}
	for _, opts := range ruleOpts {
		opts.SecGroupID = group.ID
		opts.Direction = rules.DirIngress
		opts.EtherType = rules.EtherType4
		if err := rules.Create(b.network, opts).Err; err != nil {
			return err
		}
	}

	b.instance.Status.Amphora.HealthSecurityGroupIDs = []string{group.ID}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func fetchImage(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code for %s: %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}
