package nova

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"gopkg.in/ini.v1"
	corev1 "k8s.io/api/core/v1"
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

	cfg := BuildConfig(instance, cinder)

	cm.Data["nova.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["httpd.conf"] = template.MustRenderFile(AppLabel, "httpd.conf",
		HttpdParamsFrom(8774, APIBinary, instance.Spec.API.TLS))

	cm.Data["kolla-nova-api.json"] = template.MustReadFile(AppLabel, "kolla-nova-api.json")
	cm.Data["kolla-nova-conductor.json"] = template.MustReadFile(AppLabel, "kolla-nova-conductor.json")
	cm.Data["kolla-nova-scheduler.json"] = template.MustReadFile(AppLabel, "kolla-nova-scheduler.json")

	return cm
}

func BuildConfig(instance *openstackv1beta1.Nova, cinder *openstackv1beta1.Cinder) *ini.File {
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "nova.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	if cinder != nil {
		for _, backend := range cinder.Spec.Backends {
			if cephSpec := backend.Ceph; cephSpec != nil {
				// TODO support multiple ceph backends
				cfg.Section("libvirt").NewKey("rbd_secret_uuid", "74a0b63e-041d-4040-9398-3704e4cf8260")
				cfg.Section("libvirt").NewKey("rbd_user", cephSpec.ClientName)
			}
		}
	}

	template.MergeINI(cfg, spec.ExtraConfig)

	return cfg
}

func Secret(instance *openstackv1beta1.Nova) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	secret.StringData["metadata-proxy-secret"] = template.MustGeneratePassword()

	return secret
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Nova, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Nova) {
		instance.Spec = intended.Spec
	})
}

type HttpdParams struct {
	ListenPort int32
	Binary     string
	TLS        bool
}

func HttpdParamsFrom(port int32, binary string, tlsSpec openstackv1beta1.TLSServerSpec) HttpdParams {
	return HttpdParams{
		ListenPort: port,
		Binary:     binary,
		TLS:        tlsSpec.Secret != "",
	}
}
