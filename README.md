# openstack-operator

Experimental Kubernetes operator for running OpenStack clouds at scale, built on the Operator framework.

This operator comes batteries-included, provisioning MariaDB, RabbitMQ, and other stateful services required
for a basic OpenStack deployment. In the future, it may support using externally provisioned services. The
only exception is Ceph, which depends on Rook for provisioning.

The initial goal of this operator is to allow for provisioning multiple clouds in the same Kubernetes cluster
for testing and development purposes. Production operations, also known as "day 2" operations, may come at a
later point.

The API will almost certainly have significant changes, so definitely don't use this in production.

## Images

This operator utilizes custom container images based on Ubuntu OpenStack, using the 20.04 release that ships
with [Ussuri](https://docs.openstack.org/ussuri/).

These images are built from [openstack-operator-images](https://github.com/ianunruh/openstack-operator-images)
using the Docker Hub automated build process. Currently images are just tagged as `latest`, there may be
a more stable tag in the future.

Stateful services like MariaDB, Memcached, and RabbitMQ use container images from Bitnami.

## Development

```bash
git clone https://github.com/ianunruh/openstack-operator.git
cd openstack-operator

make install
make run

# Optionally, switch to the namespace you wish to test in
kubectl config set-context --current --namespace=openstack2

kubectl apply -f config/samples/openstack_v1beta1_controlplane.yaml
```

Label the desired compute/network nodes to enable deployment.

```
kubectl label node compute7 \
    openstack/cloud=openstack2 \            ## optional if only running single cloud
    openstack/compute-node=enabled \        ## compute node must be capable of running kvm
    openstack/network-agent=linuxbridge \   ## required for both compute and network nodes
    openstack/network-node=enabled          ## compute and network can co-exist on same node
```

Use the supplied `openrc` file with the Python CLI client.

```
source openrc

openstack token issue
openstack catalog list
openstack image list
openstack resource provider list
openstack compute service list
openstack hypervisor list
openstack network agent list
openstack volume service list
```

To clean up a cluster completely, make sure the persistent volumes are deleted.

```
kubectl delete controlplane default
kubectl delete pvc -l app=mariadb
kubectl delete pvc -l app=rabbitmq
```

Ingress secrets will also remain until manually cleaned up.

## Requirements

Tested on Kubernetes 1.19 and later, but may work with earlier versions.

* cert-manager
* ingress-nginx

Optional, but recommended:

* external-dns

## Inspirations

This operator takes inspiration and some scripts from these projects.

* [openstack-helm](https://github.com/openstack/openstack-helm)
* [openstack-k8s-operators](https://github.com/openstack-k8s-operators)
