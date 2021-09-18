package rabbitmq

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ClusterComponentLabel = "cluster"
)

func ClusterStatefulSet(instance *openstackv1beta1.RabbitMQ, configHash string) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	runAsUser := int64(1001)

	// TODO pod anti-affinity
	sts := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			FSGroup:   &runAsUser,
			RunAsUser: &runAsUser,
		},
		Containers: []corev1.Container{
			{
				Name:  "rabbitmq",
				Image: instance.Spec.Image,
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"rabbitmqctl", "stop_app"},
						},
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"rabbitmq-diagnostics", "-q", "check_running"},
						},
					},
					InitialDelaySeconds: 120,
					PeriodSeconds:       30,
					TimeoutSeconds:      20,
					SuccessThreshold:    1,
					FailureThreshold:    6,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"rabbitmq-diagnostics", "-q", "ping"},
						},
					},
					InitialDelaySeconds: 10,
					PeriodSeconds:       30,
					TimeoutSeconds:      20,
					SuccessThreshold:    1,
					FailureThreshold:    3,
				},
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
					template.EnvVar("BITNAMI_DEBUG", "false"),
					template.FieldEnvVar("MY_POD_IP", "status.podIP"),
					template.FieldEnvVar("MY_POD_NAME", "metadata.name"),
					template.FieldEnvVar("MY_POD_NAMESPACE", "metadata.namespace"),
					template.EnvVar("K8S_SERVICE_NAME", "rabbitmq-headless"),
					template.EnvVar("K8S_ADDRESS_TYPE", "hostname"),
					template.EnvVar("RABBITMQ_FORCE_BOOT", "no"),
					template.EnvVar("RABBITMQ_NODE_NAME", "rabbit@$(MY_POD_NAME).$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local"),
					template.EnvVar("K8S_HOSTNAME_SUFFIX", ".$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local"),
					template.EnvVar("RABBITMQ_MNESIA_DIR", "/bitnami/rabbitmq/mnesia/$(RABBITMQ_NODE_NAME)"),
					template.EnvVar("RABBITMQ_LOGS", "-"),
					template.EnvVar("RABBITMQ_ULIMIT_NOFILES", "65536"),
					template.EnvVar("RABBITMQ_USE_LONGNAME", "true"),
					template.SecretEnvVar("RABBITMQ_ERL_COOKIE", instance.Name, "erlang-cookie"),
					template.EnvVar("RABBITMQ_USERNAME", "admin"),
					template.SecretEnvVar("RABBITMQ_PASSWORD", instance.Name, "password"),
					template.EnvVar("RABBITMQ_PLUGINS", "rabbitmq_management, rabbitmq_peer_discovery_k8s, rabbitmq_auth_backend_ldap, rabbitmq_prometheus"),
				},
				Ports: []corev1.ContainerPort{
					{Name: "amqp", ContainerPort: 5672},
					{Name: "dist", ContainerPort: 25672},
					{Name: "epmd", ContainerPort: 4369},
					{Name: "metrics", ContainerPort: 9419},
					{Name: "stats", ContainerPort: 15672},
				},
				Resources: instance.Spec.Resources,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config",
						MountPath: "/bitnami/rabbitmq/conf/rabbitmq.conf",
						SubPath:   "rabbitmq.conf",
					},
					{
						Name:      "data",
						MountPath: "/bitnami/rabbitmq/mnesia",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("config", instance.Name, nil),
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			template.PersistentVolumeClaim("data", labels, instance.Spec.Volume),
		},
	})

	sts.Spec.Template.Spec.ServiceAccountName = instance.Name

	return sts
}

func ClusterService(instance *openstackv1beta1.RabbitMQ) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	svc := template.GenericService(instance.Name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "amqp", Port: 5672},
		{Name: "dist", Port: 25672},
		{Name: "epmd", Port: 4369},
		{Name: "metrics", Port: 9419},
		{Name: "stats", Port: 15672},
	}

	return svc
}

func ClusterHeadlessService(instance *openstackv1beta1.RabbitMQ) *corev1.Service {
	svc := ClusterService(instance)
	svc.Name = template.HeadlessServiceName(instance.Name)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}

func ClusterManagementIngress(instance *openstackv1beta1.RabbitMQ) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	ingress := template.GenericIngress(instance.Name, instance.Namespace, instance.Spec.Management.Ingress, labels)

	ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Name = "stats"

	return ingress
}

func EnsureCluster(ctx context.Context, c client.Client, intended *openstackv1beta1.RabbitMQ, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.RabbitMQ{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating RabbitMQ", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating RabbitMQ", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
