package user

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/rabbitmq"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	defaultBrokerPort uint16 = 5672
	defaultAdminPort  uint16 = 15672
)

func SetupJob(instance *openstackv1beta1.RabbitMQUser) *batchv1.Job {
	labels := template.AppLabels(instance.Name, rabbitmq.AppLabel)

	spec := instance.Spec.SetupJob
	clusterName := instance.Spec.Cluster

	namePrefix := instance.Spec.Cluster

	hostname := clusterName
	port := defaultAdminPort

	adminSecret := clusterName
	adminUsernameEnv := template.EnvVar("RABBITMQ_ADMIN_USERNAME", "admin")
	secretPasswordKey := "password"

	if clusterName == "" {
		externalSpec := instance.Spec.External

		namePrefix = "external"

		hostname = externalSpec.Host
		if externalSpec.AdminPort > 0 {
			port = externalSpec.AdminPort
		}

		adminSecret = externalSpec.AdminSecret.Name
		if externalSpec.AdminSecret.UsernameKey != "" {
			adminUsernameEnv = template.SecretEnvVar("RABBITMQ_ADMIN_USERNAME", adminSecret, externalSpec.AdminSecret.UsernameKey)
		}
		if externalSpec.AdminSecret.PasswordKey != "" {
			secretPasswordKey = externalSpec.AdminSecret.PasswordKey
		}
	}

	env := []corev1.EnvVar{
		template.EnvVar("RABBIT_HOSTNAME", hostname),
		template.EnvVar("RABBIT_PORT", strconv.Itoa(int(port))),
		adminUsernameEnv,
		template.SecretEnvVar("RABBITMQ_ADMIN_PASSWORD", adminSecret, secretPasswordKey),
		template.SecretEnvVar("RABBITMQ_USER_CONNECTION", instance.Spec.Secret, "connection"),
	}

	var (
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
	)

	pki.AppendRabbitMQTLSClientVolumes(instance.Spec, &volumes, &volumeMounts)

	if instance.Spec.TLS.CABundle != "" || (instance.Spec.External != nil && instance.Spec.External.TLS.CABundle != "") {
		env = append(env, template.EnvVar("RABBITMQ_TLS_CA_BUNDLE", "/etc/ssl/certs/rabbitmq/ca.crt"))
	}

	job := template.GenericJob(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:  "user",
				Image: spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(rabbitmq.AppLabel, "user-setup.sh"),
				},
				Env:          env,
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(namePrefix, "user", instance.Name)

	return job
}

func Secret(instance *openstackv1beta1.RabbitMQUser, password string) *corev1.Secret {
	labels := template.AppLabels(instance.Name, rabbitmq.AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	clusterName := instance.Spec.Cluster

	hostname := fmt.Sprintf("%s.%s.svc", clusterName, instance.Namespace)
	port := defaultBrokerPort
	username := instance.Spec.Name
	vhost := instance.Spec.VirtualHost

	if password == "" {
		password = template.MustGeneratePassword()
	}

	useTLS := false

	if clusterName == "" {
		externalSpec := instance.Spec.External

		hostname = externalSpec.Host

		if externalSpec.TLS.CABundle != "" {
			useTLS = true
		}

		if externalSpec.Port > 0 {
			port = externalSpec.Port
		}
	} else if instance.Spec.TLS.CABundle != "" {
		useTLS = true
	}

	var driverOpts []string

	if useTLS {
		driverOpts = append(driverOpts, "ssl=True")
		driverOpts = append(driverOpts, "ssl_ca_file=/etc/ssl/certs/rabbitmq/ca.crt")
	}

	var query string
	if len(driverOpts) > 0 {
		query = fmt.Sprintf("?%s", strings.Join(driverOpts, "&"))
	}

	secret.StringData["connection"] = fmt.Sprintf(
		"rabbit://%s:%s@%s:%d/%s%s",
		username, password, hostname, port, vhost, query)

	secret.StringData["password"] = password

	return secret
}

func PasswordFromSecret(secret *corev1.Secret) string {
	return string(secret.Data["password"])
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.RabbitMQUser, log logr.Logger) error {
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.RabbitMQUser) {
		instance.Spec = intended.Spec
	})
}
