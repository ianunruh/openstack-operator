# openstack-operator

Experimental Kubernetes operator for running OpenStack clouds at scale, built on the Operator framework.

This operator comes batteries-included, provisioning MariaDB, RabbitMQ, and other stateful services required
for a basic OpenStack deployment. In the future, it may support using externally provisioned services. The
only exception is Ceph, which depends on Rook for provisioning.

The initial goal of this operator is to allow for provisioning multiple clouds in the same Kubernetes cluster
for testing and development purposes. Production operations, also known as "day 2" operations, may come at a
later point.

This operator deploys an [OVN](https://www.ovn.org/en/architecture/) control plane to provide
scalable network virtualization.

The API will almost certainly have significant changes, so definitely don't use this in production.

## Components

Supports deploying a "minimal" OpenStack cloud with the following components.

* Keystone (identity)
* Glance (image registry)
* Placement (resource allocation)
* Nova (compute)
* Neutron + OVN (networking)
* Horizon (dashboard)

Also supports the following optional components.

* Barbican (secrets)
* Cinder (block storage)
* Heat (orchestration)
* Magnum (container engine orchestration)
* Manila (filesystem storage)
* Octavia (load balancing)
* Rally (benchmarking)

## Images

This operator utilizes custom container images based on Ubuntu OpenStack, using the 20.04 release that ships
with [Wallaby](https://docs.openstack.org/wallaby/).

These images are built from [openstack-operator-images](https://github.com/ianunruh/openstack-operator-images)
using GitHub Actions for CI. Currently images are just tagged as `latest`, there may be
a more stable tag in the future.

Stateful services like MariaDB, Memcached, and RabbitMQ use container images from Bitnami.

## Custom resources

Check out the [Godoc reference](https://pkg.go.dev/github.com/ianunruh/openstack-operator/api/v1beta1#ControlPlaneSpec) or
the [samples](https://github.com/ianunruh/openstack-operator/tree/master/config/samples).

## Development

```bash
git clone https://github.com/ianunruh/openstack-operator.git
cd openstack-operator

make install
make run

# Optionally, switch to the namespace you wish to test in
kubectl config set-context --current --namespace=openstack

kubectl apply -f config/samples/openstack_v1beta1_controlplane.yaml
```

Label the desired compute/network nodes to enable deployment.

```
kubectl label node compute7 \
    openstack/cloud=openstack \           ## optional if only running single cloud
    openstack/compute-node=enabled \      ## compute node must be capable of running KVM
    openstack/network-agent=ovn           ## only one network agent supported currently
```

Use the supplied `openrc` file with the Python CLI client.

```
pip install osc-placement python-heatclient python-magnumclient python-manilaclient python-openstackclient

source openrc

openstack token issue
openstack catalog list
openstack image list
openstack resource provider list
openstack compute service list
openstack hypervisor list
openstack network agent list
openstack volume service list
openstack orchestration service list
openstack coe service list
```

Launch an instance

```
source openrc

openstack flavor create m1.test --ram 1024 --disk 1

curl -OL http://download.cirros-cloud.net/0.5.1/cirros-0.5.1-x86_64-disk.img
openstack image create cirros --public --disk-format qcow2 --file cirros-0.5.1-x86_64-disk.img

openstack network create test1
openstack subnet create test1 --network test1 --subnet-range 10.91.0.0/24

openstack router create test1
openstack router add subnet test1 test1

NET_ID=$(openstack network show test1 -f json | jq -r .id)
openstack server create test1 --image cirros --flavor m1.test --network $NET_ID
```

Visit the OpenStack dashboard in your browser

```
# Visit this hostname with /horizon appended
kubectl get horizon horizon -o 'jsonpath={.spec.server.ingress.host}'

# The domain is "default" and the admin password can be retrieved with
kubectl get secret keystone -o 'jsonpath={.data.OS_PASSWORD}' | base64 -d
```

To clean up a cluster completely, make sure the persistent volumes are deleted.

```
kubectl delete controlplane default

kubectl delete pvc --all
# alternatively target specific PVCs
kubectl delete pvc -l app=mariadb
kubectl delete pvc -l app=memcached
kubectl delete pvc -l app=rabbitmq
```

Ingress secrets will also remain until manually cleaned up.

Compute hosts need additional cleanup. It's recommended that hosts are rebooted
after tearing down the cluster, as there will be leftover networking state, such
as network namespaces, iptables rules, and OVS data paths.

```
pkill -f "OpenStack Nova"
rm -rf /var/lib/nova /var/lib/libvirt /var/run/openvswitch /var/lib/openvswitch
```

## Requirements

Tested on Kubernetes 1.21 and later, but may work with earlier versions.

* cert-manager
* ingress-nginx

Optional, but recommended:

* external-dns

## Inspirations

This operator takes inspiration and some scripts from these projects.

* [openstack-helm](https://github.com/openstack/openstack-helm)
* [openstack-k8s-operators](https://github.com/openstack-k8s-operators)
