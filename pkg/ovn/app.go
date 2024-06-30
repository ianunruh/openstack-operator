package ovn

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "ovn"
)

func ConfigMap(instance *openstackv1beta1.OVNControlPlane) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["get-encap-ip.py"] = template.MustReadFile(AppLabel, "get-encap-ip.py")
	cm.Data["setup-node.sh"] = template.MustReadFile(AppLabel, "setup-node.sh")
	cm.Data["start-northd.sh"] = template.MustReadFile(AppLabel, "start-northd.sh")
	cm.Data["start-ovsdb-nb.sh"] = template.MustReadFile(AppLabel, "start-ovsdb-nb.sh")
	cm.Data["start-ovsdb-sb.sh"] = template.MustReadFile(AppLabel, "start-ovsdb-sb.sh")

	cm.Data["kolla-controller.json"] = template.MustReadFile(AppLabel, "kolla-controller.json")
	cm.Data["kolla-northd.json"] = template.MustReadFile(AppLabel, "kolla-northd.json")
	cm.Data["kolla-openvswitch-ovsdb.json"] = template.MustReadFile(AppLabel, "kolla-openvswitch-ovsdb.json")
	cm.Data["kolla-openvswitch-vswitchd.json"] = template.MustReadFile(AppLabel, "kolla-openvswitch-vswitchd.json")
	cm.Data["kolla-ovsdb.json"] = template.MustReadFile(AppLabel, "kolla-ovsdb.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.OVNControlPlane, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.OVNControlPlane) {
		instance.Spec = intended.Spec
	})
}
