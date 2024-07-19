package rabbitmq

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/prometheus"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ClusterComponentLabel = "cluster"
)

func ClusterStatefulSet(instance *openstackv1beta1.RabbitMQ, configHash string) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	runAsUser := int64(1001)

	env := []corev1.EnvVar{
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
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("config", "/bitnami/rabbitmq/conf/rabbitmq.conf", "rabbitmq.conf"),
		template.VolumeMount("data", "/bitnami/rabbitmq/mnesia"),
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("config", instance.Name, nil),
	}

	pki.AppendTLSServerVolumes(instance.Spec.TLS, "/opt/bitnami/rabbitmq/certs", 0444, &volumes, &volumeMounts)

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
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"rabbitmqctl", "stop_app"},
						},
					},
				},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
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
					ProbeHandler: corev1.ProbeHandler{
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
				Env: env,
				Ports: []corev1.ContainerPort{
					{Name: "amqp", ContainerPort: 5672},
					{Name: "dist", ContainerPort: 25672},
					{Name: "epmd", ContainerPort: 4369},
					{Name: "metrics", ContainerPort: 9419},
					{Name: "stats", ContainerPort: 15672},
				},
				Resources:    instance.Spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
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
		{Name: "stats", Port: 15672},
	}

	return svc
}

func ClusterHeadlessService(instance *openstackv1beta1.RabbitMQ) *corev1.Service {
	extraPorts := []corev1.ServicePort{
		{Name: "metrics", Port: 9419},
	}

	svc := ClusterService(instance)
	svc.Name = template.HeadlessServiceName(instance.Name)
	svc.Spec.ClusterIP = corev1.ClusterIPNone
	svc.Spec.Ports = append(svc.Spec.Ports, extraPorts...)

	return svc
}

func ClusterManagementIngress(instance *openstackv1beta1.RabbitMQ) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	ingress := template.GenericIngressWithTLS(
		instance.Name,
		instance.Namespace,
		instance.Spec.Management.Ingress,
		instance.Spec.TLS,
		labels)

	ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Name = "stats"

	return ingress
}

func ClusterServiceMonitor(instance *openstackv1beta1.RabbitMQ) *unstructured.Unstructured {
	return prometheus.ServiceMonitor(prometheus.ServiceMonitorParams{
		Name:          instance.Name,
		Namespace:     instance.Namespace,
		NameLabel:     AppLabel,
		InstanceLabel: instance.Name,
	})
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.RabbitMQ, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.RabbitMQ) {
		instance.Spec = intended.Spec
	})
}
