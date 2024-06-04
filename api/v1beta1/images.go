package v1beta1

const (
	DefaultBarbicanAPIImage    = "kolla/barbican-api:2023.2-ubuntu-jammy"
	DefaultBarbicanWorkerImage = "kolla/barbican-worker:2023.2-ubuntu-jammy"

	DefaultCinderAPIImage       = "kolla/cinder-api:2023.2-ubuntu-jammy"
	DefaultCinderSchedulerImage = "kolla/cinder-scheduler:2023.2-ubuntu-jammy"
	DefaultCinderVolumeImage    = "kolla/cinder-volume:2023.2-ubuntu-jammy"

	DefaultGlanceAPIImage = "kolla/glance-api:2023.2-ubuntu-jammy"

	DefaultHeatAPIImage    = "kolla/heat-api:2023.2-ubuntu-jammy"
	DefaultHeatCFNImage    = "kolla/heat-api-cfn:2023.2-ubuntu-jammy"
	DefaultHeatEngineImage = "kolla/heat-engine:2023.2-ubuntu-jammy"

	DefaultHorizonServerImage = "kolla/horizon:2023.2-ubuntu-jammy"

	DefaultKeystoneAPIImage = "kolla/keystone:2023.2-ubuntu-jammy"

	DefaultMagnumAPIImage       = "kolla/magnum-api:2023.2-ubuntu-jammy"
	DefaultMagnumConductorImage = "kolla/magnum-conductor:2023.2-ubuntu-jammy"

	DefaultManilaAPIImage       = "kolla/manila-api:2023.2-ubuntu-jammy"
	DefaultManilaSchedulerImage = "kolla/manila-scheduler:2023.2-ubuntu-jammy"
	DefaultManilaShareImage     = "kolla/manila-share:2023.2-ubuntu-jammy"

	DefaultNeutronMetadataAgentImage = "kolla/neutron-metadata-agent:2023.2-ubuntu-jammy"
	DefaultNeutronServerImage        = "kolla/neutron-server:2023.2-ubuntu-jammy"

	DefaultNovaAPIImage       = "kolla/nova-api:2023.2-ubuntu-jammy"
	DefaultNovaConductorImage = "kolla/nova-conductor:2023.2-ubuntu-jammy"
	DefaultNovaSchedulerImage = "kolla/nova-scheduler:2023.2-ubuntu-jammy"

	DefaultNovaNoVNCProxyImage = "kolla/nova-novncproxy:2023.2-ubuntu-jammy"

	DefaultNovaComputeImage    = "kolla/nova-compute:2023.2-ubuntu-jammy"
	DefaultNovaComputeSSHImage = "kolla/nova-ssh:2023.2-ubuntu-jammy"
	DefaultNovaLibvirtdImage   = "kolla/nova-libvirt:2023.2-ubuntu-jammy"

	DefaultOctaviaAPIImage           = "kolla/octavia-api:2023.2-ubuntu-jammy"
	DefaultOctaviaDriverAgentImage   = "kolla/octavia-driver-agent:2023.2-ubuntu-jammy"
	DefaultOctaviaHealthManagerImage = "kolla/octavia-health-manager:2023.2-ubuntu-jammy"
	DefaultOctaviaHousekeepingImage  = "kolla/octavia-housekeeping:2023.2-ubuntu-jammy"
	DefaultOctaviaWorkerImage        = "kolla/octavia-worker:2023.2-ubuntu-jammy"

	DefaultOVNControllerImage = "kolla/ovn-controller:2023.2-ubuntu-jammy"
	DefaultOVNNorthdImage     = "kolla/ovn-northd:2023.2-ubuntu-jammy"
	DefaultOVNOVSDBNorthImage = "kolla/ovn-nb-db-server:2023.2-ubuntu-jammy"
	DefaultOVNOVSDBSouthImage = "kolla/ovn-sb-db-server:2023.2-ubuntu-jammy"

	DefaultOVSDBImage     = "kolla/openvswitch-db-server:2023.2-ubuntu-jammy"
	DefaultOVSSwitchImage = "kolla/openvswitch-vswitchd:2023.2-ubuntu-jammy"

	DefaultPlacementAPIImage = "kolla/placement-api:2023.2-ubuntu-jammy"

	DefaultRallyImage = "xrally/xrally-openstack:2.1.0"

	DefaultMariaDBImage         = "docker.io/bitnami/mariadb:10.5.8-debian-10-r21"
	DefaultMariaDBExporterImage = "docker.io/bitnami/mysqld-exporter:0.13.0"

	DefaultMemcachedImage         = "docker.io/bitnami/memcached:1.6.9-debian-10-r0"
	DefaultMemcachedExporterImage = "docker.io/bitnami/memcached-exporter:0.9.0"

	DefaultRabbitMQImage           = "docker.io/bitnami/rabbitmq:3.8.9-debian-10-r58"
	DefaultRabbitMQManagementImage = "rabbitmq:3.8.9-management"
)

func imageDefault(image, fallback string) string {
	if image == "" {
		return fallback
	}
	return image
}
