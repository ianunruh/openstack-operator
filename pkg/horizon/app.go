package horizon

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "horizon"
)

var (
	appUID = int64(42420)
)

func ConfigMap(instance *openstackv1beta1.Horizon) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["local_settings.py"] = template.MustRenderFile(AppLabel, "local_settings.py", configParamsFrom(instance))

	return cm
}

type configParams struct {
	SSO configSSOParams
}

type configSSOParams struct {
	Enabled       bool
	KeystoneURL   string
	InitialChoice string
	Choices       []configSSOChoice
}

type configSSOChoice struct {
	Kind  string
	Title string
}

func configParamsFrom(instance *openstackv1beta1.Horizon) configParams {
	params := configParams{}

	if ssoSpec := instance.Spec.SSO; ssoSpec.Enabled {
		var (
			initialChoice string
			choices       []configSSOChoice
		)

		for _, method := range ssoSpec.Methods {
			if method.Default {
				initialChoice = method.Kind
			}
			choices = append(choices, configSSOChoice{
				Kind:  method.Kind,
				Title: method.Title,
			})
		}

		params.SSO = configSSOParams{
			Enabled:       true,
			KeystoneURL:   ssoSpec.KeystoneURL,
			InitialChoice: initialChoice,
			Choices:       choices,
		}
	}

	return params
}

func EnsureHorizon(ctx context.Context, c client.Client, intended *openstackv1beta1.Horizon, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Horizon{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Horizon", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Horizon", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
