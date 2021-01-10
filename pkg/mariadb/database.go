package mariadb

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

	hostname := instance.Spec.Cluster
	username := instance.Spec.Name
	password := template.NewPassword()
	db := instance.Spec.Name

	secret.StringData["connection"] = fmt.Sprintf("mysql+pymysql://%s:%s@%s:3306/%s", username, password, hostname, db)
	secret.StringData["password"] = password

	return secret
}

func EnsureDatabase(ctx context.Context, c client.Client, instance *openstackv1beta1.MariaDBDatabase, log logr.Logger) error {
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

		log.Info("Creating MariaDBDatabase", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating MariaDBDatabase", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
