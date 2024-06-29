package keystone

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
	AppLabel = "keystone"
)

var (
	appUID = int64(42425)
)

func ConfigMap(instance *openstackv1beta1.Keystone) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "keystone.conf")

	cfg.Section("cache").NewKey("backend_argument",
		fmt.Sprintf("url:%s", strings.Join(spec.Cache.Servers, ",")))

	if spec.Notifications.Enabled {
		cfg.Section("oslo_messaging_notifications").NewKey("driver", "messagingv2")
	}

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["httpd.conf"] = template.MustReadFile(AppLabel, "httpd.conf")
	cm.Data["kolla.json"] = template.MustReadFile(AppLabel, "kolla.json")

	cm.Data["keystone.conf"] = template.MustOutputINI(cfg).String()

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Keystone, log logr.Logger) error {
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(instance, hash)

		log.Info("Creating Keystone", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		template.SetAppliedHash(instance, hash)

		log.Info("Updating Keystone", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
