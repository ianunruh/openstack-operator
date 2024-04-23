package v1beta1

const (
	DefaultBarbicanAPIImage     = "kolla/barbican-api:2023.2-ubuntu-jammy"
	DefaultBarbicanWorkerImage  = "kolla/barbican-worker:2023.2-ubuntu-jammy"
	DefaultCinderAPIImage       = "kolla/cinder-api:2023.2-ubuntu-jammy"
	DefaultCinderSchedulerImage = "kolla/cinder-scheduler:2023.2-ubuntu-jammy"
	DefaultCinderVolumeImage    = "kolla/cinder-volume:2023.2-ubuntu-jammy"
	DefaultGlanceAPIImage       = "kolla/glance-api:2023.2-ubuntu-jammy"
	DefaultHeatAPIImage         = "kolla/heat-api:2023.2-ubuntu-jammy"
	DefaultHeatCFNImage         = "kolla/heat-api-cfn:2023.2-ubuntu-jammy"
	DefaultHeatEngineImage      = "kolla/heat-engine:2023.2-ubuntu-jammy"
	DefaultHorizonServerImage   = "kolla/horizon:2023.2-ubuntu-jammy"
	DefaultKeystoneAPIImage     = "kolla/keystone:2023.2-ubuntu-jammy"
	DefaultLibvirtImage         = "ghcr.io/ianunruh/openstack-operator-images/libvirt:master"
	DefaultMagnumAPIImage       = "kolla/magnum-api:2023.2-ubuntu-jammy"
	DefaultMagnumConductorImage = "kolla/magnum-conductor:2023.2-ubuntu-jammy"
	DefaultManilaAPIImage       = "kolla/manila-api:2023.2-ubuntu-jammy"
	DefaultManilaSchedulerImage = "kolla/manila-scheduler:2023.2-ubuntu-jammy"
	DefaultManilaShareImage     = "kolla/manila-share:2023.2-ubuntu-jammy"
	DefaultNeutronImage         = "ghcr.io/ianunruh/openstack-operator-images/neutron:master"
	DefaultNovaImage            = "ghcr.io/ianunruh/openstack-operator-images/nova:master"
	DefaultOctaviaImage         = "ghcr.io/ianunruh/openstack-operator-images/octavia:master"
	DefaultPlacementAPIImage    = "kolla/placement-api:2023.2-ubuntu-jammy"

	DefaultRallyImage = "xrally/xrally-openstack:2.1.0"

	DefaultMariaDBImage         = "docker.io/bitnami/mariadb:10.5.8-debian-10-r21"
	DefaultMariaDBExporterImage = "docker.io/bitnami/mysqld-exporter:0.13.0"

	DefaultMemcachedImage         = "docker.io/bitnami/memcached:1.6.9-debian-10-r0"
	DefaultMemcachedExporterImage = "docker.io/bitnami/memcached-exporter:0.9.0"

	DefaultRabbitMQImage = "docker.io/bitnami/rabbitmq:3.8.9-debian-10-r58"
)

func imageDefault(image, fallback string) string {
	if image == "" {
		return fallback
	}
	return image
}
