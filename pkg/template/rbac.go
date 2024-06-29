package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenericServiceAccount(name, namespace string, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func GenericRole(name, namespace string, labels map[string]string, rules []rbacv1.PolicyRule) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: rules,
	}
}

func GenericRoleBinding(name, namespace string, labels map[string]string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func RoleRef(name string) rbacv1.RoleRef {
	return rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     name,
	}
}

func EnsureServiceAccount(ctx context.Context, c client.Client, instance *corev1.ServiceAccount, log logr.Logger) error {
	hash, err := ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating ServiceAccount", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.Secrets = intended.Secrets
		instance.ImagePullSecrets = intended.ImagePullSecrets

		SetAppliedHash(instance, hash)

		log.Info("Updating ServiceAccount", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

func EnsureRole(ctx context.Context, c client.Client, instance *rbacv1.Role, log logr.Logger) error {
	hash, err := ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating Role", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.Rules = intended.Rules

		SetAppliedHash(instance, hash)

		log.Info("Updating Role", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

func EnsureRoleBinding(ctx context.Context, c client.Client, instance *rbacv1.RoleBinding, log logr.Logger) error {
	hash, err := ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(instance, hash)

		log.Info("Creating RoleBinding", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		instance.RoleRef = intended.RoleRef
		instance.Subjects = intended.Subjects

		SetAppliedHash(instance, hash)

		log.Info("Updating RoleBinding", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
