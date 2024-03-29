package controlplane

const (
	DefaultRabbitMQImage  = "docker.io/bitnami/rabbitmq:3.8.9-debian-10-r58"
	DefaultMemcachedImage = "docker.io/bitnami/memcached:1.6.9-debian-10-r0"
	DefaultMariaDBImage   = "docker.io/bitnami/mariadb:10.5.8-debian-10-r21"
	DefaultRallyImage     = "xrally/xrally-openstack:2.1.0"

	DefaultBarbicanImage  = "kolla/barbican-api:2023.2-ubuntu-jammy"
	DefaultCinderImage    = "ghcr.io/ianunruh/openstack-operator-images/cinder:master"
	DefaultGlanceImage    = "kolla/glance-api:2023.2-ubuntu-jammy"
	DefaultHeatImage      = "kolla/heat-api:2023.2-ubuntu-jammy"
	DefaultHorizonImage   = "kolla/horizon:2023.2-ubuntu-jammy"
	DefaultKeystoneImage  = "kolla/keystone:2023.2-ubuntu-jammy"
	DefaultLibvirtImage   = "ghcr.io/ianunruh/openstack-operator-images/libvirt:master"
	DefaultMagnumImage    = "ghcr.io/ianunruh/openstack-operator-images/magnum:master"
	DefaultManilaImage    = "ghcr.io/ianunruh/openstack-operator-images/manila:master"
	DefaultNeutronImage   = "ghcr.io/ianunruh/openstack-operator-images/neutron:master"
	DefaultNovaImage      = "ghcr.io/ianunruh/openstack-operator-images/nova:master"
	DefaultOctaviaImage   = "ghcr.io/ianunruh/openstack-operator-images/octavia:master"
	DefaultPlacementImage = "kolla/placement-api:2023.2-ubuntu-jammy"
	DefaultSenlinImage    = "kolla/senlin-api:2023.2-ubuntu-jammy"
)

func imageDefault(image, fallback string) string {
	if image == "" {
		return fallback
	}
	return image
}
