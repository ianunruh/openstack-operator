package amphora

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	amphoraImageName   = "amphora"
	amphoraImageTag    = "amphora"
	amphoraKeypairName = "amphora"
	amphoraNetworkName = "octavia-lb-mgmt"

	imageSourceProperty = "source"

	healthPortDeviceOwner = "Octavia:health-mgr"
	healthPortPrefix      = "octavia-health-manager-"
)

func Bootstrap(ctx context.Context, instance *openstackv1beta1.Octavia, c client.Client, report template.ReportFunc, log logr.Logger) (ctrl.Result, error) {
	b, err := newBootstrap(ctx, instance, c, log)
	if err != nil {
		if err := report(ctx, "Error during amphora bootstrap: %v", err); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if err := b.EnsureAll(ctx); err != nil {
		if err := report(ctx, "Error during amphora bootstrap: %v", err); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	return b.Wait(ctx, report)
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
		deps:     template.NewConditionWaiter(c.Scheme(), log),
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

	if err := b.EnsureHealthPorts(ctx); err != nil {
		return err
	}

	return nil
}

func (b *bootstrap) Wait(ctx context.Context, report template.ReportFunc) (ctrl.Result, error) {
	return b.deps.Wait(ctx, report)
}
