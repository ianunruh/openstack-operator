package senlin

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
	AppLabel = "senlin"
)

var (
	appUID = int64(42443)
)

func ConfigMap(instance *openstackv1beta1.Senlin) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINI(AppLabel, "senlin.conf")
	template.MergeINI(cfg, instance.Spec.ExtraConfig)

	cm.Data["senlin.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-senlin-api.json"] = template.MustReadFile(AppLabel, "kolla-senlin-api.json")
	cm.Data["kolla-senlin-conductor.json"] = template.MustReadFile(AppLabel, "kolla-senlin-conductor.json")
	cm.Data["kolla-senlin-engine.json"] = template.MustReadFile(AppLabel, "kolla-senlin-engine.json")
	cm.Data["kolla-senlin-health-manager.json"] = template.MustReadFile(AppLabel, "kolla-senlin-health-manager.json")

	return cm
}

func EnsureSenlin(ctx context.Context, c client.Client, intended *openstackv1beta1.Senlin, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Senlin{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Senlin", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Senlin", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
