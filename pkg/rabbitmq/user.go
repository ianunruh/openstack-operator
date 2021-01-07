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
				Name:  "user",
				Image: containerImage,
				Command: []string{
					"bash",
					"-c",
					template.MustRenderFile(AppLabel, "user.sh", nil),
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

	secret.StringData["connection"] = fmt.Sprintf("rabbitmq://%s:%s@%s/%s", username, password, hostname, vhost)
	secret.StringData["password"] = password

	return secret
}

func EnsureUser(ctx context.Context, c client.Client, intended *openstackv1beta1.RabbitMQUser, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.RabbitMQUser{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating RabbitMQUser", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating RabbitMQUser", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
