package rally

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "rally"
)

func ConfigMap(instance *openstackv1beta1.Rally) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "rally.conf")

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["rally.conf"] = template.MustOutputINI(cfg).String()

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Rally, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Rally) {
		instance.Spec = intended.Spec
	})
}
