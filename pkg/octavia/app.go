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
	spec := instance.Spec

	cfg := template.MustLoadINI(AppLabel, "octavia.conf")

	cfg.Section("keystone_authtoken").NewKey("memcached_servers", strings.Join(spec.Cache.Servers, ","))

	var providerDrivers []string
	if spec.Amphora.Enabled {
		providerDrivers = append(providerDrivers, "amphora:The Octavia Amphora driver")
	}
	if spec.OVN.Enabled {
		providerDrivers = append(providerDrivers, "ovn:Octavia OVN driver")
	}

	cfg.Section("api_settings").NewKey("enabled_provider_drivers", strings.Join(providerDrivers, ","))

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

	template.MergeINI(cfg, spec.ExtraConfig)

	cm.Data["octavia.conf"] = template.MustOutputINI(cfg).String()

	cm.Data["httpd.conf"] = template.MustReadFile(AppLabel, "httpd.conf")

	cm.Data["kolla-octavia-api.json"] = template.MustReadFile(AppLabel, "kolla-octavia-api.json")
	cm.Data["kolla-octavia-driver-agent.json"] = template.MustReadFile(AppLabel, "kolla-octavia-driver-agent.json")
	cm.Data["kolla-octavia-health-manager.json"] = template.MustReadFile(AppLabel, "kolla-octavia-health-manager.json")
	cm.Data["kolla-octavia-housekeeping.json"] = template.MustReadFile(AppLabel, "kolla-octavia-housekeeping.json")
	cm.Data["kolla-octavia-worker.json"] = template.MustReadFile(AppLabel, "kolla-octavia-worker.json")

	return cm
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.Octavia, log logr.Logger) error {
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

		log.Info("Creating Octavia", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		template.SetAppliedHash(instance, hash)

		log.Info("Updating Octavia", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}
