package neutron

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
	AppLabel = "neutron"
)

var (
	appUID = int64(42435)
)

func ConfigMap(instance *openstackv1beta1.Neutron) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "neutron.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["neutron.conf"] = template.MustOutputINI(cfg).String()
	cm.Data["neutron_ovn_metadata_agent.ini"] = template.MustReadFile(AppLabel, "neutron_ovn_metadata_agent.ini")

	cm.Data["kolla-neutron-metadata-agent.json"] = template.MustReadFile(AppLabel, "kolla-neutron-metadata-agent.json")
	cm.Data["kolla-neutron-server.json"] = template.MustReadFile(AppLabel, "kolla-neutron-server.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Neutron, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Neutron) {
		instance.Spec = intended.Spec
	})
}
