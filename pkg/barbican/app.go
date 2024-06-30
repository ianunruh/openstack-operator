package barbican

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "barbican"
)

var (
	appUID = int64(42403)
)

func ConfigMap(instance *openstackv1beta1.Barbican) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "barbican.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["barbican.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-barbican-api.json"] = template.MustReadFile(AppLabel, "kolla-barbican-api.json")
	cm.Data["kolla-barbican-worker.json"] = template.MustReadFile(AppLabel, "kolla-barbican-worker.json")

	return cm
}

func Secret(instance *openstackv1beta1.Barbican) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	secret.StringData["kek"] = template.MustGenerateFernetKey()

	return secret
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Barbican, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Barbican) {
		instance.Spec = intended.Spec
	})
}
