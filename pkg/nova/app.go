package nova

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
	AppLabel = "nova"
)

var (
	appUID = int64(42436)
)

func ConfigMap(instance *openstackv1beta1.Nova, cinder *openstackv1beta1.Cinder) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINI(AppLabel, "nova.conf")

	if cinder != nil {
		for _, backend := range cinder.Spec.Backends {
			if cephSpec := backend.Ceph; cephSpec != nil {
				// TODO support multiple ceph backends
				cfg.Section("libvirt").NewKey("rbd_secret_uuid", "74a0b63e-041d-4040-9398-3704e4cf8260")
				cfg.Section("libvirt").NewKey("rbd_user", cephSpec.ClientName)
			}
		}
	}

	template.MergeINI(cfg, instance.Spec.ExtraConfig)

	cm.Data["nova.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-nova-api.json"] = template.MustReadFile(AppLabel, "kolla-nova-api.json")
	cm.Data["kolla-nova-conductor.json"] = template.MustReadFile(AppLabel, "kolla-nova-conductor.json")
	cm.Data["kolla-nova-novncproxy.json"] = template.MustReadFile(AppLabel, "kolla-nova-novncproxy.json")
	cm.Data["kolla-nova-scheduler.json"] = template.MustReadFile(AppLabel, "kolla-nova-scheduler.json")

	return cm
}

func Secret(instance *openstackv1beta1.Nova) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	secret.StringData["metadata-proxy-secret"] = template.NewPassword()

	return secret
}

func EnsureNova(ctx context.Context, c client.Client, intended *openstackv1beta1.Nova, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Nova{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Nova", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Nova", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
