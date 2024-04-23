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

	cfg := template.MustLoadINI(AppLabel, "manila.conf")

	var backendNames []string

	for _, backend := range instance.Spec.Backends {
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

	template.MergeINI(cfg, instance.Spec.ExtraConfig)

	cm.Data["manila.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-manila-api.json"] = template.MustReadFile(AppLabel, "kolla-manila-api.json")
	cm.Data["kolla-manila-scheduler.json"] = template.MustReadFile(AppLabel, "kolla-manila-scheduler.json")
	cm.Data["kolla-manila-share.json"] = template.MustReadFile(AppLabel, "kolla-manila-share.json")

	return cm
}

func EnsureManila(ctx context.Context, c client.Client, intended *openstackv1beta1.Manila, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Manila{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Manila", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Manila", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
