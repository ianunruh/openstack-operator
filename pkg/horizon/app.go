package horizon

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "horizon"
)

func ConfigMap(instance *openstackv1beta1.Horizon) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["httpd.conf"] = template.MustReadFile(AppLabel, "httpd.conf")
	cm.Data["kolla.json"] = template.MustReadFile(AppLabel, "kolla.json")
	cm.Data["local_settings.py"] = template.MustReadFile(AppLabel, "local_settings.py")

	return cm
}

func Secret(instance *openstackv1beta1.Horizon) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Name, instance.Namespace, labels)

	secret.StringData["secret-key"] = template.MustGeneratePassword()

	return secret
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Horizon, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.Horizon) {
		instance.Spec = intended.Spec
	})
}
