package glance

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
	AppLabel = "glance"
)

var (
	appUID = int64(42415)
)

func ConfigMap(instance *openstackv1beta1.Glance) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "glance-api.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	var (
		backendNames   []string
		defaultBackend string
	)

	for _, backend := range spec.Backends {
		var backendType string

		section := cfg.Section(backend.Name)

		if cephSpec := backend.Ceph; cephSpec != nil {
			backendType = "rbd"

			section.NewKey("rbd_store_pool", cephSpec.PoolName)
			section.NewKey("rbd_store_user", cephSpec.ClientName)
			section.NewKey("rbd_store_ceph_conf", filepath.Join("/etc/ceph", cephSpec.Secret, "ceph.conf"))

			// TODO if cinder has a ceph backend, then enable this
			// cfg.Section("").NewKey("show_image_direct_url", "true")
		} else if pvcSpec := backend.PVC; pvcSpec != nil {
			backendType = "file"

			section.NewKey("filesystem_store_datadir", imageBackendPath(backend.Name))
		}

		backendNames = append(backendNames, fmt.Sprintf("%s:%s", backend.Name, backendType))

		if backend.Default || defaultBackend == "" {
			defaultBackend = backend.Name
		}
	}

	cfg.Section("").NewKey("enabled_backends", strings.Join(backendNames, ","))
	cfg.Section("glance_store").NewKey("default_backend", defaultBackend)

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["glance-api.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla.json"] = template.MustReadFile(AppLabel, "kolla.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Glance, log logr.Logger) error {
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

		log.Info("Creating Glance", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		template.SetAppliedHash(instance, hash)

		log.Info("Updating Glance", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

func imageBackendPath(name string) string {
	return template.Combine("/var/lib/glance/images", name)
}
