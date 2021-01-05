package mariadb

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

type databaseOptions struct {
	DatabaseName          string
	DatabaseHostname      string
	DatabaseAdminUsername string
}

func DatabaseJob(instance *openstackv1beta1.MariaDBDatabase, containerImage, databaseHostName, adminSecret string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	opts := databaseOptions{
		DatabaseName:          instance.Spec.Name,
		DatabaseHostname:      databaseHostName,
		DatabaseAdminUsername: "root",
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "database",
				Image: containerImage,
				Command: []string{
					"bash",
					"-c",
					template.MustRenderFile(AppLabel, "database.sh", opts),
				},
				Env: []corev1.EnvVar{
					template.SecretEnvVar("MYSQL_PWD", adminSecret, "password"),
					template.SecretEnvVar("DATABASE_PASSWORD", instance.Spec.Secret, "password"),
				},
			},
		},
	})

	job.Name = template.Combine(instance.Spec.Cluster, "database", instance.Name)

	return job
}

func DatabaseSecret(instance *openstackv1beta1.MariaDBDatabase) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	secret.StringData["password"] = template.NewPassword()

	return secret
}

func EnsureDatabase(ctx context.Context, c client.Client, intended *openstackv1beta1.MariaDBDatabase, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.MariaDBDatabase{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating MariaDB database", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating MariaDB database", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
