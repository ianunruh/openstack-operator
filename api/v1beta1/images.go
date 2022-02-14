package v1beta1

const (
	BarbicanDefaultImage  = "ghcr.io/ianunruh/openstack-operator-images/barbican:master"
	CinderDefaultImage    = "ghcr.io/ianunruh/openstack-operator-images/cinder:master"
	GlanceDefaultImage    = "ghcr.io/ianunruh/openstack-operator-images/glance:master"
	HeatDefaultImage      = "ghcr.io/ianunruh/openstack-operator-images/heat:master"
	HorizonDefaultImage   = "ghcr.io/ianunruh/openstack-operator-images/horizon:master"
	KeystoneDefaultImage  = "ghcr.io/ianunruh/openstack-operator-images/keystone:master"
	LibvirtDefaultImage   = "ghcr.io/ianunruh/openstack-operator-images/libvirt:master"
	MagnumDefaultImage    = "ghcr.io/ianunruh/openstack-operator-images/magnum:master"
	ManilaDefaultImage    = "ghcr.io/ianunruh/openstack-operator-images/manila:master"
	NeutronDefaultImage   = "ghcr.io/ianunruh/openstack-operator-images/neutron:master"
	NovaDefaultImage      = "ghcr.io/ianunruh/openstack-operator-images/nova:master"
	OctaviaDefaultImage   = "ghcr.io/ianunruh/openstack-operator-images/octavia:master"
	PlacementDefaultImage = "ghcr.io/ianunruh/openstack-operator-images/placement:master"
	SenlinDefaultImage    = "ghcr.io/ianunruh/openstack-operator-images/senlin:master"

	RallyDefaultImage = "xrally/xrally-openstack:2.1.0"

	MariaDBDefaultImage         = "docker.io/bitnami/mariadb:10.5.8-debian-10-r21"
	MariaDBExporterDefaultImage = "docker.io/bitnami/mysqld-exporter:0.13.0"

	MemcachedDefaultImage         = "docker.io/bitnami/memcached:1.6.9-debian-10-r0"
	MemcachedExporterDefaultImage = "docker.io/bitnami/memcached-exporter:0.9.0"

	RabbitMQDefaultImage = "docker.io/bitnami/rabbitmq:3.8.9-debian-10-r58"
)

func imageDefault(image, imageDefault string) string {
	if image == "" {
		return imageDefault
	}
	return image
}
