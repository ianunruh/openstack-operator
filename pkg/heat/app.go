package heat

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "heat"
)

var (
	appUID = int64(42418)
)

func ConfigMap(instance *openstackv1beta1.Heat) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "heat.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["heat.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-heat-api.json"] = template.MustReadFile(AppLabel, "kolla-heat-api.json")
	cm.Data["kolla-heat-api-cfn.json"] = template.MustReadFile(AppLabel, "kolla-heat-api-cfn.json")
	cm.Data["kolla-heat-engine.json"] = template.MustReadFile(AppLabel, "kolla-heat-engine.json")

	return cm
}

func EnsureHeat(ctx context.Context, c client.Client, intended *openstackv1beta1.Heat, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Heat{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Heat", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Heat", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
