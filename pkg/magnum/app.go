package magnum

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
	AppLabel = "magnum"
)

var (
	appUID = int64(42428)
)

func ConfigMap(instance *openstackv1beta1.Magnum) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "magnum.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["magnum.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-magnum-api.json"] = template.MustReadFile(AppLabel, "kolla-magnum-api.json")
	cm.Data["kolla-magnum-conductor.json"] = template.MustReadFile(AppLabel, "kolla-magnum-conductor.json")

	return cm
}

func EnsureMagnum(ctx context.Context, c client.Client, intended *openstackv1beta1.Magnum, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Magnum{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Magnum", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Magnum", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
