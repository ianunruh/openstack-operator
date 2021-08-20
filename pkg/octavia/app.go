package octavia

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	AppLabel = "octavia"
)

var (
	appUID = int64(42437)
)

func ConfigMap(instance *openstackv1beta1.Octavia) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cfg := template.MustLoadINI(AppLabel, "octavia.conf")

	amphora := instance.Status.Amphora

	var healthManagerAddrs []string
	for _, port := range amphora.HealthPorts {
		healthManagerAddrs = append(healthManagerAddrs, fmt.Sprintf("%s:5555", port.IPAddress))
	}

	cfg.Section("controller_worker").NewKey("amp_flavor_id", amphora.FlavorID)
	cfg.Section("controller_worker").NewKey("amp_image_owner_id", amphora.ImageProjectID)
	cfg.Section("controller_worker").NewKey("amp_secgroup_list", strings.Join(amphora.SecurityGroupIDs, ","))
	cfg.Section("controller_worker").NewKey("amp_boot_network_list", strings.Join(amphora.NetworkIDs, ","))

	cfg.Section("health_manager").NewKey("controller_ip_port_list", strings.Join(healthManagerAddrs, ","))

	cm.Data["octavia.conf"] = template.MustOutputINI(cfg).String()

	return cm
}

func EnsureOctavia(ctx context.Context, c client.Client, intended *openstackv1beta1.Octavia, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Octavia{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Octavia", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Octavia", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
