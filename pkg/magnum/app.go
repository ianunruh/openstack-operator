package magnum

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
	AppLabel = "magnum"
)

var (
	appUID = int64(42428)
)

func ConfigMap(instance *openstackv1beta1.Magnum) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "magnum.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["magnum.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["kolla-magnum-api.json"] = template.MustReadFile(AppLabel, "kolla-magnum-api.json")
	cm.Data["kolla-magnum-conductor.json"] = template.MustReadFile(AppLabel, "kolla-magnum-conductor.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Magnum, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Magnum) {
		instance.Spec = intended.Spec
	})
}
