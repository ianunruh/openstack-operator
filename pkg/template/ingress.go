package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnsureIngress(ctx context.Context, c client.Client, intended *netv1.Ingress, log logr.Logger) error {
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &netv1.Ingress{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(intended, hash)

		log.Info("Creating Ingress", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		SetAppliedHash(found, hash)

		log.Info("Updating Ingress", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

func IngressServiceBackend(svcName, portName string) netv1.IngressBackend {
	return netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: svcName,
			Port: netv1.ServiceBackendPort{
				Name: portName,
			},
		},
	}
}
