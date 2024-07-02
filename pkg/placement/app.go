package placement

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "placement"
)

var (
	appUID = int64(42482)
)

func ConfigMap(instance *openstackv1beta1.Placement) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "placement.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	pki.SetupKeystoneMiddleware(cfg, spec.TLS)

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["placement.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["httpd.conf"] = template.MustReadFile(AppLabel, "httpd.conf")
	cm.Data["kolla.json"] = template.MustReadFile(AppLabel, "kolla.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Placement, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Placement) {
		instance.Spec = intended.Spec
	})
}
