package user

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
	"github.com/gophercloud/utils/openstack/clientconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Reconcile(instance *openstackv1beta1.KeystoneUser, secret *corev1.Secret, identity *gophercloud.ServiceClient, log logr.Logger) error {
	project, err := reconcileProject(instance.Spec.Project, instance.Spec.ProjectDomain, identity, log)
	if err != nil {
		return err
	}

	domain, err := reconcileDomain(instance.Spec.Domain, identity, log)
	if err != nil {
		return err
	}

	name := instance.Spec.Name
	if name == "" {
		name = instance.Name
	}

	password := string(secret.Data["OS_PASSWORD"])

	user, err := findUserByName(name, domain.ID, identity)
	if err != nil {
		return err
	}

	if user == nil {
		opts := users.CreateOpts{
			DomainID: domain.ID,
			Name:     name,
			Password: password,
		}
		if project != nil {
			opts.DefaultProjectID = project.ID
		}

		log.Info("Creating user", "name", name)
		user, err = users.Create(identity, opts).Extract()
		if err != nil {
			return err
		}
	} else {
		opts := users.UpdateOpts{
			DomainID: domain.ID,
			Password: password,
		}
		if project != nil {
			opts.DefaultProjectID = project.ID
		}

		log.Info("Updating user", "name", name)
		if err := users.Update(identity, user.ID, opts).Err; err != nil {
			return err
		}
	}

	rolesToAssign, err := reconcileRoles(instance.Spec.Roles, identity, log)
	if err != nil {
		return err
	}

	if err := reconcileUserRoles(user, project, rolesToAssign, identity, log); err != nil {
		return err
	}

	return nil
}

func reconcileDomain(name string, identity *gophercloud.ServiceClient, log logr.Logger) (*domains.Domain, error) {
	domain, err := findDomainByName(name, identity)
	if err != nil {
		return nil, err
	}

	if domain == nil {
		log.Info("Creating domain", "name", name)
		domain, err = domains.Create(identity, domains.CreateOpts{
			Name: name,
		}).Extract()
		if err != nil {
			return nil, err
		}
	}

	return domain, nil
}

func reconcileProject(name, domainName string, identity *gophercloud.ServiceClient, log logr.Logger) (*projects.Project, error) {
	if name == "" {
		return nil, nil
	}

	domain, err := reconcileDomain(domainName, identity, log)
	if err != nil {
		return nil, err
	}

	project, err := findProjectByName(name, domain.ID, identity)
	if err != nil {
		return nil, err
	}

	if project == nil {
		log.Info("Creating project", "name", name)
		project, err = projects.Create(identity, projects.CreateOpts{
			DomainID: domain.ID,
			Name:     name,
		}).Extract()
		if err != nil {
			return nil, err
		}
	}

	return project, nil
}

func reconcileRoles(names []string, identity *gophercloud.ServiceClient, log logr.Logger) ([]*roles.Role, error) {
	pages, err := roles.List(identity, roles.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("listing roles: %w", err)
	}

	current, err := roles.ExtractRoles(pages)
	if err != nil {
		return nil, fmt.Errorf("extracting roles: %w", err)
	}

	filtered := make([]*roles.Role, 0, len(names))
	for _, name := range names {
		role, err := reconcileRole(name, current, identity, log)
		if err != nil {
			return nil, err
		}

		filtered = append(filtered, role)
	}

	return filtered, nil
}

func reconcileRole(name string, current []roles.Role, identity *gophercloud.ServiceClient, log logr.Logger) (*roles.Role, error) {
	role := filterRoleByName(name, current)

	if role == nil {
		log.Info("Creating role", "name", name)
		var err error
		role, err = roles.Create(identity, roles.CreateOpts{
			Name: name,
		}).Extract()
		if err != nil {
			return nil, err
		}
	}

	return role, nil
}

func reconcileUserRoles(user *users.User, project *projects.Project, rolesToAssign []*roles.Role, identity *gophercloud.ServiceClient, log logr.Logger) error {
	opts := roles.AssignOpts{
		UserID: user.ID,
	}
	if project == nil {
		opts.DomainID = user.DomainID
	} else {
		opts.ProjectID = project.ID
	}

	for _, role := range rolesToAssign {
		if err := roles.Assign(identity, role.ID, opts).Err; err != nil {
			return err
		}
	}

	return nil
}

func filterRoleByName(name string, current []roles.Role) *roles.Role {
	for _, role := range current {
		if role.Name == name {
			return &role
		}
	}

	return nil
}

func findDomainByName(name string, identity *gophercloud.ServiceClient) (*domains.Domain, error) {
	pages, err := domains.List(identity, domains.ListOpts{
		Name: name,
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("listing domains: %w", err)
	}

	current, err := domains.ExtractDomains(pages)
	if err != nil {
		return nil, fmt.Errorf("extracting domains: %w", err)
	}

	if len(current) == 0 {
		return nil, nil
	}

	return &current[0], nil
}

func findProjectByName(name, domainID string, identity *gophercloud.ServiceClient) (*projects.Project, error) {
	pages, err := projects.List(identity, projects.ListOpts{
		DomainID: domainID,
		Name:     name,
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	current, err := projects.ExtractProjects(pages)
	if err != nil {
		return nil, fmt.Errorf("extracting projects: %w", err)
	}

	if len(current) == 0 {
		return nil, nil
	}

	return &current[0], nil
}

func findUserByName(name, domainID string, identity *gophercloud.ServiceClient) (*users.User, error) {
	pages, err := users.List(identity, users.ListOpts{
		DomainID: domainID,
		Name:     name,
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	current, err := users.ExtractUsers(pages)
	if err != nil {
		return nil, fmt.Errorf("extracting users: %w", err)
	}

	if len(current) == 0 {
		return nil, nil
	}

	return &current[0], nil
}

func Secret(instance *openstackv1beta1.KeystoneUser, cluster *openstackv1beta1.Keystone, password string) *corev1.Secret {
	labels := template.AppLabels(instance.Name, keystone.AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	username := instance.Spec.Name
	if username == "" {
		username = instance.Name
	}

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
		password = template.MustGeneratePassword()
	}

	cloudsYAML := clientconfig.Clouds{
		Clouds: map[string]clientconfig.Cloud{
			"default": {
				AuthInfo: &clientconfig.AuthInfo{
					AuthURL:           wwwAuthURL,
					Username:          username,
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
		"OS_USERNAME":             username,
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
