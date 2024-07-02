package cinder

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "cinder"
)

var (
	appUID = int64(42407)
)

func ConfigMap(instance *openstackv1beta1.Cinder) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "cinder.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	var backendNames []string

	for _, backend := range spec.Backends {
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

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["cinder.conf"] = template.MustOutputINI(cfg).String()
	cm.Data["httpd.conf"] = template.MustRenderFile(AppLabel, "httpd.conf", httpdParamsFrom(instance))

	cm.Data["kolla-cinder-api.json"] = template.MustReadFile(AppLabel, "kolla-cinder-api.json")
	cm.Data["kolla-cinder-backup.json"] = template.MustReadFile(AppLabel, "kolla-cinder-backup.json")
	cm.Data["kolla-cinder-scheduler.json"] = template.MustReadFile(AppLabel, "kolla-cinder-scheduler.json")
	cm.Data["kolla-cinder-volume.json"] = template.MustReadFile(AppLabel, "kolla-cinder-volume.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Cinder, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Cinder) {
		instance.Spec = intended.Spec
	})
}

type httpdParams struct {
	TLS bool
}

func httpdParamsFrom(instance *openstackv1beta1.Cinder) httpdParams {
	return httpdParams{
		TLS: instance.Spec.API.TLS.Secret != "",
	}
}
