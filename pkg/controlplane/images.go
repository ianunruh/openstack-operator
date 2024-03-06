package controlplane

const (
	DefaultRabbitMQImage  = "docker.io/bitnami/rabbitmq:3.8.9-debian-10-r58"
	DefaultMemcachedImage = "docker.io/bitnami/memcached:1.6.9-debian-10-r0"
	DefaultMariaDBImage   = "docker.io/bitnami/mariadb:10.5.8-debian-10-r21"
	DefaultRallyImage     = "xrally/xrally-openstack:2.1.0"

	DefaultBarbicanImage  = "ghcr.io/ianunruh/openstack-operator-images/barbican:master"
	DefaultCinderImage    = "ghcr.io/ianunruh/openstack-operator-images/cinder:master"
	DefaultGlanceImage    = "ghcr.io/ianunruh/openstack-operator-images/glance:master"
	DefaultHeatImage      = "ghcr.io/ianunruh/openstack-operator-images/heat:master"
	DefaultHorizonImage   = "kolla/horizon:2023.2-ubuntu-jammy"
	DefaultKeystoneImage  = "ghcr.io/ianunruh/openstack-operator-images/keystone:master"
	DefaultLibvirtImage   = "ghcr.io/ianunruh/openstack-operator-images/libvirt:master"
	DefaultMagnumImage    = "ghcr.io/ianunruh/openstack-operator-images/magnum:master"
	DefaultManilaImage    = "ghcr.io/ianunruh/openstack-operator-images/manila:master"
	DefaultNeutronImage   = "ghcr.io/ianunruh/openstack-operator-images/neutron:master"
	DefaultNovaImage      = "ghcr.io/ianunruh/openstack-operator-images/nova:master"
	DefaultOctaviaImage   = "ghcr.io/ianunruh/openstack-operator-images/octavia:master"
	DefaultPlacementImage = "ghcr.io/ianunruh/openstack-operator-images/placement:master"
	DefaultSenlinImage    = "ghcr.io/ianunruh/openstack-operator-images/senlin:master"
)

func imageDefault(image, fallback string) string {
	if image == "" {
		return fallback
	}
	return image
}
