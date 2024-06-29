package database

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
	"github.com/ianunruh/openstack-operator/pkg/mariadb"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const defaultPort uint16 = 3306

func SetupJob(instance *openstackv1beta1.MariaDBDatabase) *batchv1.Job {
	labels := template.AppLabels(instance.Name, mariadb.AppLabel)

	spec := instance.Spec.SetupJob
	clusterName := instance.Spec.Cluster

	namePrefix := clusterName

	hostname := clusterName
	port := defaultPort

	adminSecret := clusterName
	secretPasswordKey := "password"

	if clusterName == "" {
		externalSpec := instance.Spec.External

		namePrefix = "external"

		hostname = externalSpec.Host
		if externalSpec.Port > 0 {
			port = externalSpec.Port
		}

		adminSecret = externalSpec.AdminSecret.Name
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
				Name:  "database",
				Image: spec.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(mariadb.AppLabel, "database.sh"),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("MYSQL_HOST", hostname),
					template.EnvVar("MYSQL_TCP_PORT", strconv.Itoa(int(port))),
					template.SecretEnvVar("MYSQL_PWD", adminSecret, secretPasswordKey),
					template.EnvVar("DATABASE_ADMIN_USER", "root"),
					template.EnvVar("DATABASE_NAME", instance.Spec.Name),
					template.SecretEnvVar("DATABASE_PASSWORD", instance.Spec.Secret, "password"),
				},
				Resources: spec.Resources,
			},
		},
	})

	job.Name = template.Combine(namePrefix, "database", instance.Name)

	return job
}

func Secret(instance *openstackv1beta1.MariaDBDatabase) *corev1.Secret {
	labels := template.AppLabels(instance.Name, mariadb.AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	clusterName := instance.Spec.Cluster

	hostname := clusterName
	username := instance.Spec.Name
	password := template.MustGeneratePassword()
	db := instance.Spec.Name
	port := defaultPort

	if clusterName == "" {
		externalSpec := instance.Spec.External

		hostname = externalSpec.Host
		if externalSpec.Port > 0 {
			port = externalSpec.Port
		}
	}

	secret.StringData["connection"] = fmt.Sprintf("mysql+pymysql://%s:%s@%s:%d/%s", username, password, hostname, port, db)
	secret.StringData["password"] = password

	return secret
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.MariaDBDatabase, log logr.Logger) error {
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
