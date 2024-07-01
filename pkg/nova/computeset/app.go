package computeset

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func ConfigMap(instance *openstackv1beta1.NovaComputeSet, cell *openstackv1beta1.NovaCell, cluster *openstackv1beta1.Nova, cinder *openstackv1beta1.Cinder) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, nova.AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := nova.BuildConfig(cluster, cinder)

	template.MergeINI(cfg, cell.Spec.ExtraConfig)
	template.MergeINI(cfg, instance.Spec.ExtraConfig)

	cm.Data["nova.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["compute-ssh.sh"] = template.MustReadFile(nova.AppLabel, "compute-ssh.sh")

	cm.Data["kolla-nova-compute.json"] = template.MustReadFile(nova.AppLabel, "kolla-nova-compute.json")
	cm.Data["kolla-nova-compute-ssh.json"] = template.MustReadFile(nova.AppLabel, "kolla-nova-compute-ssh.json")

	return cm
}
