package manila

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "manila"
)

var (
	appUID = int64(42429)
)

func ConfigMap(instance *openstackv1beta1.Manila) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "manila.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	var backendNames []string

	for _, backend := range spec.Backends {
		section := cfg.Section(backend.Name)
		section.NewKey("share_backend_name", backend.ShareBackendName)

		if cephSpec := backend.Ceph; cephSpec != nil {
			section.NewKey("driver_handles_share_servers", "false")
			section.NewKey("share_driver", "manila.share.drivers.cephfs.driver.CephFSDriver")
			section.NewKey("cephfs_conf_path", filepath.Join("/etc/ceph", cephSpec.Secret, "ceph.conf"))
			section.NewKey("cephfs_auth_id", cephSpec.ClientName)
			section.NewKey("cephfs_enable_snapshots", "true")
		}

		backendNames = append(backendNames, backend.Name)
	}

	cfg.Section("").NewKey("enabled_share_backends", strings.Join(backendNames, ","))

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["manila.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-manila-api.json"] = template.MustReadFile(AppLabel, "kolla-manila-api.json")
	cm.Data["kolla-manila-scheduler.json"] = template.MustReadFile(AppLabel, "kolla-manila-scheduler.json")
	cm.Data["kolla-manila-share.json"] = template.MustReadFile(AppLabel, "kolla-manila-share.json")

	return cm
}

func EnsureManila(ctx context.Context, c client.Client, instance *openstackv1beta1.Manila, log logr.Logger) error {
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

		log.Info("Creating Manila", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		template.SetAppliedHash(instance, hash)

		log.Info("Updating Manila", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
