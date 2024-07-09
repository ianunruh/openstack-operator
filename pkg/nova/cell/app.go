package cell

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func ConfigMap(instance *openstackv1beta1.NovaCell, cluster *openstackv1beta1.Nova, cinder *openstackv1beta1.Cinder) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, nova.AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := nova.BuildConfig(cluster, cinder)

	template.MergeINI(cfg, instance.Spec.ExtraConfig)

	cm.Data["nova.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["httpd.conf"] = template.MustRenderFile(nova.AppLabel, "httpd.conf", nova.HttpdParamsFrom(8774, instance.Spec.Metadata.TLS))

	cm.Data["kolla-nova-api.json"] = template.MustReadFile(nova.AppLabel, "kolla-nova-api.json")
	cm.Data["kolla-nova-conductor.json"] = template.MustReadFile(nova.AppLabel, "kolla-nova-conductor.json")
	cm.Data["kolla-nova-novncproxy.json"] = template.MustReadFile(nova.AppLabel, "kolla-nova-novncproxy.json")

	return cm
}
