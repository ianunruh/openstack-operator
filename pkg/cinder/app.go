package cinder

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
	AppLabel = "cinder"
)

var (
	appUID = int64(64061)
)

func ConfigMap(instance *openstackv1beta1.Cinder) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINITemplate(AppLabel, "cinder.conf", nil)

	var backendNames []string

	for _, backend := range instance.Spec.Backends {
		section := cfg.Section(backend.Name)
		section.NewKey("volume_backend_name", backend.VolumeBackendName)

		if cephSpec := backend.Ceph; cephSpec != nil {
			section.NewKey("rbd_ceph_conf", filepath.Join("/etc/ceph", cephSpec.Secret, "ceph.conf"))
			// TODO support multiple secret UUIDs
			section.NewKey("rbd_secret_uuid", "74a0b63e-041d-4040-9398-3704e4cf8260")
			section.NewKey("rbd_pool", cephSpec.PoolName)
			section.NewKey("rbd_user", cephSpec.ClientName)
			section.NewKey("report_discard_supported", "true")
			section.NewKey("volume_driver", "cinder.volume.drivers.rbd.RBDDriver")
		}

		backendNames = append(backendNames, backend.Name)
	}

	cfg.Section("").NewKey("enabled_backends", strings.Join(backendNames, ","))

	cm.Data["cinder.conf"] = template.MustOutputINI(cfg).String()

	return cm
}

func EnsureCinder(ctx context.Context, c client.Client, intended *openstackv1beta1.Cinder, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Cinder{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Cinder", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Cinder", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
