package neutron

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
	AppLabel = "neutron"
)

func ConfigMap(instance *openstackv1beta1.Neutron) *corev1.ConfigMap {
	labels := template.AppLabels(instance.Name, AppLabel)
	cm := template.GenericConfigMap(instance.Name, instance.Namespace, labels)

	cm.Data["dhcp_agent.ini"] = template.MustReadFile(AppLabel, "dhcp_agent.ini")
	cm.Data["l3_agent.ini"] = template.MustReadFile(AppLabel, "l3_agent.ini")
	cm.Data["linuxbridge_agent.ini"] = linuxBridgeAgentCfg(instance.Spec.LinuxBridgeAgent)
	cm.Data["metadata_agent.ini"] = template.MustReadFile(AppLabel, "metadata_agent.ini")
	cm.Data["ml2_conf.ini"] = template.MustReadFile(AppLabel, "ml2_conf.ini")
	cm.Data["neutron.conf"] = template.MustReadFile(AppLabel, "neutron.conf")

	return cm
}

func linuxBridgeAgentCfg(spec openstackv1beta1.NeutronLinuxBridgeAgentSpec) string {
	cfg := template.MustLoadINI(AppLabel, "linuxbridge_agent.ini")

	physicalInterfaceMappings := strings.Join(spec.PhysicalInterfaceMappings, ",")
	cfg.Section("linux_bridge").NewKey("physical_interface_mappings", physicalInterfaceMappings)

	return template.MustOutputINI(cfg).String()
}

func EnsureNeutron(ctx context.Context, c client.Client, intended *openstackv1beta1.Neutron, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Neutron{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Neutron", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Neutron", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}
