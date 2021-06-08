package rabbitmq

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func UserJob(instance *openstackv1beta1.RabbitMQUser, containerImage, databaseHostName, adminSecret string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name: "user",
				// TODO make configurable
				Image: "rabbitmq:3.8.9-management",
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(AppLabel, "user-setup.sh"),
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("RABBITMQ_ADMIN_CONNECTION", adminSecret, "connection"),
					template.SecretEnvVar("RABBITMQ_USER_CONNECTION", instance.Spec.Secret, "connection"),
				},
			},
		},
	})

	job.Name = template.Combine(instance.Spec.Cluster, "user", instance.Name)

	return job
}

func UserSecret(instance *openstackv1beta1.RabbitMQUser) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	hostname := instance.Spec.Cluster
	username := instance.Spec.Name
	password := template.NewPassword()
	vhost := instance.Spec.VirtualHost

	secret.StringData["connection"] = fmt.Sprintf("rabbit://%s:%s@%s:5672/%s", username, password, hostname, vhost)
	secret.StringData["password"] = password

	return secret
}

func EnsureUser(ctx context.Context, c client.Client, instance *openstackv1beta1.RabbitMQUser, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

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
