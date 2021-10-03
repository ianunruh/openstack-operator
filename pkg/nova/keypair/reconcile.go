package keypair

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Reconcile(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaKeypair, compute *gophercloud.ServiceClient, identity *gophercloud.ServiceClient, log logr.Logger) error {
	userID, err := getUserID(instance, identity, log)
	if err != nil {
		return err
	}

	keypair, err := getKeypair(instance, userID, compute)
	if err != nil {
		return err
	}

	if err := reconcileKeypair(ctx, c, instance, keypair, userID, compute, log); err != nil {
		return err
	}

	return nil
}

func getKeypair(instance *openstackv1beta1.NovaKeypair, userID string, compute *gophercloud.ServiceClient) (*keypairs.KeyPair, error) {
	name := keypairName(instance)
	return keypairs.Get(compute, name, keypairs.GetOpts{
		UserID: userID,
	}).Extract()
}

func reconcileKeypair(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaKeypair, keypair *keypairs.KeyPair, userID string, compute *gophercloud.ServiceClient, log logr.Logger) error {
	name := keypairName(instance)

	var err error

	// create new keypair
	if keypair == nil {
		log.Info("Creating host keypair", "name", name)
		keypair, err = keypairs.Create(compute, keypairs.CreateOpts{
			Name:      name,
			PublicKey: instance.Spec.PublicKey,
			UserID:    userID,
		}).Extract()
		if err != nil {
			return err
		}
	}

	// TODO replace keypair if public key changed

	return nil
}

func Delete(instance *openstackv1beta1.NovaKeypair, compute *gophercloud.ServiceClient, identity *gophercloud.ServiceClient, log logr.Logger) error {
	name := keypairName(instance)

	userID, err := getUserID(instance, identity, log)
	if err != nil {
		return err
	}

	log.Info("Deleting host keypair", "name", name)
	if err := keypairs.Delete(compute, name, keypairs.DeleteOpts{
		UserID: userID,
	}).Err; err != nil {
		if errors.Is(err, gophercloud.ErrDefault404{}) {
			log.Info("Keypair not found on deletion", "name", name)
		} else {
			return err
		}
	}

	return nil
}

func keypairName(instance *openstackv1beta1.NovaKeypair) string {
	if instance.Spec.Name == "" {
		return instance.Name
	}
	return instance.Spec.Name
}

func getUserID(instance *openstackv1beta1.NovaKeypair, identity *gophercloud.ServiceClient, log logr.Logger) (string, error) {
	if instance.Spec.User == "" {
		return "", nil
	}

	domainID, err := getDomainID(instance, identity, log)
	if err != nil {
		return "", nil
	}

	pages, err := users.List(identity, users.ListOpts{
		Name:     instance.Spec.User,
		DomainID: domainID,
	}).AllPages()
	if err != nil {
		return "", err
	}

	current, err := users.ExtractUsers(pages)
	if err != nil {
		return "", err
	}

	if len(current) == 0 {
		return "", fmt.Errorf("user not found: %s", instance.Spec.User)
	}

	return current[0].ID, nil
}

func getDomainID(instance *openstackv1beta1.NovaKeypair, identity *gophercloud.ServiceClient, log logr.Logger) (string, error) {
	if instance.Spec.UserDomain == "" {
		return "", nil
	}

	pages, err := domains.List(identity, domains.ListOpts{
		Name: instance.Spec.UserDomain,
	}).AllPages()
	if err != nil {
		return "", err
	}

	current, err := domains.ExtractDomains(pages)
	if err != nil {
		return "", err
	}

	if len(current) == 0 {
		return "", fmt.Errorf("domain not found: %s", instance.Spec.UserDomain)
	}

	return current[0].ID, nil
}
