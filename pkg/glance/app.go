package glance

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "glance"
)

func ConfigMap(instance *openstackv1beta1.Glance) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINITemplate(AppLabel, "glance-api.conf", nil)

	cephSpec := instance.Spec.Storage.RookCeph
	if cephSpec != nil {
		cfg.Section("").NewKey("enabled_backends", "ceph:rbd")
		// cfg.Section("").NewKey("show_image_direct_url", "true")

		cfg.Section("glance_store").NewKey("default_backend", "ceph")

		cfg.Section("ceph").NewKey("rbd_store_pool", cephSpec.PoolName)
		cfg.Section("ceph").NewKey("rbd_store_user", cephSpec.ClientName)
	}

	cm.Data["glance-api.conf"] = template.MustOutputINI(cfg).String()

	return cm
}

func EnsureGlance(ctx context.Context, c client.Client, intended *openstackv1beta1.Glance, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Glance{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Glance", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Glance", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
