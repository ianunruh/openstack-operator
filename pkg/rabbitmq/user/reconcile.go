package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
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
				Env: []corev1.EnvVar{
					template.EnvVar("RABBIT_HOSTNAME", hostname),
					template.EnvVar("RABBIT_PORT", strconv.Itoa(int(port))),
					adminUsernameEnv,
					template.SecretEnvVar("RABBITMQ_ADMIN_PASSWORD", adminSecret, secretPasswordKey),
					template.SecretEnvVar("RABBITMQ_USER_CONNECTION", instance.Spec.Secret, "connection"),
				},
				Resources: spec.Resources,
			},
		},
	})

	job.Name = template.Combine(namePrefix, "user", instance.Name)

	return job
}

func Secret(instance *openstackv1beta1.RabbitMQUser) *corev1.Secret {
	labels := template.AppLabels(instance.Name, rabbitmq.AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	clusterName := instance.Spec.Cluster

	hostname := clusterName
	port := defaultBrokerPort
	username := instance.Spec.Name
	password := template.MustGeneratePassword()
	vhost := instance.Spec.VirtualHost

	if clusterName == "" {
		externalSpec := instance.Spec.External

		hostname = externalSpec.Host
		if externalSpec.Port > 0 {
			port = externalSpec.Port
		}
	}

	secret.StringData["connection"] = fmt.Sprintf("rabbit://%s:%s@%s:%d/%s", username, password, hostname, port, vhost)
	secret.StringData["password"] = password

	return secret
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.RabbitMQUser, log logr.Logger) error {
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(instance, hash)

		log.Info("Creating RabbitMQUser", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating RabbitMQUser", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
