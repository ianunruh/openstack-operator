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
* Senlin (clustering)

## Images

This operator utilizes custom container images based on Ubuntu OpenStack, using the 22.04 Jammy release that ships
with [OpenStack 2023.2](https://docs.openstack.org/2023.2/).

These images are built from [openstack-operator-images](https://github.com/ianunruh/openstack-operator-images)
using GitHub Actions for CI. Currently images are just tagged as `master`, there may be
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

Label the desired compute nodes to enable deployment.

```
kubectl label node compute7 \
    openstack/cloud=openstack \
    openstack/compute-node=enabled
```

Use the supplied `openrc` file with the Python CLI client.

```
# this runs in the current context's namespace
source openrc

# starts an interactive shell
pipenv install
pipenv run openstack

# the following commands run in the interactive shell

token issue
catalog list
image list
resource provider list
compute service list
hypervisor list
network agent list
volume service list
orchestration service list
coe service list
```

Some APIs require using the system scope for administrative tasks.

```
# Should be done in a new session
source openrc-system

# starts an interactive shell
pipenv run openstack

# the following commands run in the interactive shell

loadbalancer list
loadbalancer amphora list
share service list
```

Launch an instance

```
source openrc

pipenv shell

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
kubectl delete pvc -l app.kubernetes.io/name=mariadb
kubectl delete pvc -l app.kubernetes.io/name=memcached
kubectl delete pvc -l app.kubernetes.io/name=rabbitmq
kubectl delete pvc -l app.kubernetes.io/name=ovn
```

Ingress secrets will also remain until manually cleaned up.

Compute hosts need additional cleanup. It's recommended that hosts are rebooted
after tearing down the cluster, as there will be leftover networking state, such
as network namespaces, iptables rules, and OVS data paths.

```
pkill -f "OpenStack Nova"
rm -rf /var/lib/nova /etc/libvirt /var/lib/libvirt
rm -rf /etc/openvswitch /var/lib/openvswitch /var/run/openvswitch
```

## Requirements

Tested on Kubernetes 1.26, but may work with other versions.

* cert-manager
* ingress-nginx

Optional, but recommended:

* external-dns

## Inspirations

This operator takes inspiration and some scripts from these projects.

* [openstack-helm](https://github.com/openstack/openstack-helm)
* [openstack-k8s-operators](https://github.com/openstack-k8s-operators)
