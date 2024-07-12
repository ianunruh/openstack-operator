package neutron

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki/tlsproxy"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "neutron"
)

var (
	appUID int64 = 42435
)

func ConfigMap(instance *openstackv1beta1.Neutron) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "neutron.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	if spec.Server.TLS.Secret != "" {
		cfg.Section("").NewKey("bind_host", "127.0.0.1")
	}

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["neutron.conf"] = template.MustOutputINI(cfg).String()
	cm.Data["neutron_ovn_metadata_agent.ini"] = renderMetadataConfig(instance)

	cm.Data["kolla-neutron-metadata-agent.json"] = template.MustReadFile(AppLabel, "kolla-neutron-metadata-agent.json")
	cm.Data["kolla-neutron-server.json"] = template.MustReadFile(AppLabel, "kolla-neutron-server.json")

	cm.Data["tlsproxy.conf"] = tlsproxy.MustReadConfig()

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Neutron, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Neutron) {
		instance.Spec = intended.Spec
	})
}

func renderMetadataConfig(instance *openstackv1beta1.Neutron) string {
	cfg := template.MustLoadINI(AppLabel, "neutron_ovn_metadata_agent.ini")

	cfg.Section("").NewKey("nova_metadata_host", instance.Spec.Nova.MetadataHost)
	cfg.Section("").NewKey("nova_metadata_protocol", instance.Spec.Nova.MetadataProtocol)

	template.MergeINI(cfg, instance.Spec.MetadataAgent.ExtraConfig)

	return template.MustOutputINI(cfg).String()
}
