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

func ConfigMap(instance *openstackv1beta1.Glance) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINITemplate(AppLabel, "glance-api.conf", nil)

	var (
		backendNames   []string
		defaultBackend string
	)

	for _, backend := range instance.Spec.Backends {
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

func imageBackendPath(name string) string {
	return template.Combine("/var/lib/glance/images", name)
}
