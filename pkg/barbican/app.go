package barbican

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
	AppLabel = "barbican"
)

var (
	appUID = int64(42403)
)

func ConfigMap(instance *openstackv1beta1.Barbican) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINI(AppLabel, "barbican.conf")
	template.MergeINI(cfg, instance.Spec.ExtraConfig)

	cm.Data["barbican.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-barbican-api.json"] = template.MustReadFile(AppLabel, "kolla-barbican-api.json")
	cm.Data["kolla-barbican-worker.json"] = template.MustReadFile(AppLabel, "kolla-barbican-worker.json")

	return cm
}

func Secret(instance *openstackv1beta1.Barbican) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	secret.StringData["kek"] = template.MustGenerateFernetKey()

	return secret
}

func EnsureBarbican(ctx context.Context, c client.Client, intended *openstackv1beta1.Barbican, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Barbican{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Barbican", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Barbican", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
