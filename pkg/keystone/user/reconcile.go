package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gophercloud/utils/openstack/clientconfig"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func SetupJob(instance *openstackv1beta1.KeystoneUser, containerImage, adminSecret string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, keystone.AppLabel)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "setup",
				Image: containerImage,
				Command: []string{
					"python3",
					"-c",
					template.MustReadFile(keystone.AppLabel, "user-setup.py"),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("SVC_ROLES", strings.Join(instance.Spec.Roles, ",")),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(adminSecret),
					template.EnvFromSecretPrefixed(instance.Spec.Secret, "SVC_"),
				},
			},
		},
	})

	job.Name = template.Combine("keystone", "user", instance.Name)

	return job
}

func Secret(instance *openstackv1beta1.KeystoneUser, cluster *openstackv1beta1.Keystone, password string) *corev1.Secret {
	labels := template.AppLabels(instance.Name, keystone.AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	domainName := instance.Spec.Domain
	if domainName == "" {
		domainName = "Default"
	}

	projectDomainName := instance.Spec.ProjectDomain
	if projectDomainName == "" {
		projectDomainName = domainName
	}

	authURL := fmt.Sprintf("http://%s-api.%s.svc:5000/v3", cluster.Name, cluster.Namespace)

	wwwAuthURL := authURL
	if cluster.Spec.API.Ingress != nil {
		wwwAuthURL = fmt.Sprintf("https://%s/v3", cluster.Spec.API.Ingress.Host)
	}

	if password == "" {
		password = template.NewPassword()
	}

	cloudsYAML := clientconfig.Clouds{
		Clouds: map[string]clientconfig.Cloud{
			"default": {
				AuthInfo: &clientconfig.AuthInfo{
					AuthURL:           wwwAuthURL,
					Username:          instance.Name,
					Password:          password,
					ProjectName:       instance.Spec.Project,
					ProjectDomainName: projectDomainName,
					UserDomainName:    domainName,
				},
				RegionName: "RegionOne",
			},
		},
	}

	secret.StringData = map[string]string{
		"OS_IDENTITY_API_VERSION": "3",
		"OS_AUTH_URL":             authURL,
		"OS_AUTH_URL_WWW":         wwwAuthURL,
		"OS_REGION_NAME":          "RegionOne",
		"OS_PROJECT_DOMAIN_NAME":  projectDomainName,
		"OS_USER_DOMAIN_NAME":     domainName,
		"OS_PROJECT_NAME":         instance.Spec.Project,
		"OS_USERNAME":             instance.Name,
		"OS_PASSWORD":             password,
		"clouds.yaml":             string(template.MustEncodeYAML(cloudsYAML)),
	}

	return secret
}

func PasswordFromSecret(secret *corev1.Secret) string {
	return string(secret.Data["OS_PASSWORD"])
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.KeystoneUser, log logr.Logger) error {
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

		log.Info("Creating KeystoneUser", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating KeystoneUser", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
