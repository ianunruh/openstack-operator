#!/bin/bash
set -ex

if [ -n "$(cat /proc/*/comm 2>/dev/null | grep -w libvirtd)" ]; then
  set +x
  for proc in $(ls /proc/*/comm 2>/dev/null); do
    if [ "x$(cat $proc 2>/dev/null | grep -w libvirtd)" == "xlibvirtd" ]; then
      set -x
      libvirtpid=$(echo $proc | cut -f 3 -d '/')
      echo "WARNING: libvirtd daemon already running on host" 1>&2
      echo "$(cat "/proc/${libvirtpid}/status" 2>/dev/null | grep State)" 1>&2
      kill -9 "$libvirtpid" || true
      set +x
    fi
  done
  set -x
fi

rm -f /var/run/libvirtd.pid

# if [[ -c /dev/kvm ]]; then
#     chmod 660 /dev/kvm
#     chown root:kvm /dev/kvm
# fi

#Setup Cgroups to use when breaking out of Kubernetes defined groups
CGROUPS=""
for CGROUP in cpu rdma hugetlb; do
  if [ -d /sys/fs/cgroup/${CGROUP} ]; then
    CGROUPS+="${CGROUP},"
  fi
done
cgcreate -g ${CGROUPS%,}:/osh-libvirt

# Start libvirtd in background
cgexec -g ${CGROUPS%,}:/osh-libvirt systemd-run --scope --slice=system libvirtd --listen &

# Wait until libvirtd is up
TIMEOUT=60
while [[ ! -f /var/run/libvirtd.pid ]]; do
  if [[ ${TIMEOUT} -gt 0 ]]; then
    let TIMEOUT-=1
    sleep 1
  else
    echo "ERROR: libvirt did not start in time (pid file missing)"
    exit 1
  fi
done

# Even though we see the pid file the socket immediately (this is
# needed for virsh)
TIMEOUT=10
while [[ ! -e /var/run/libvirt/libvirt-sock ]]; do
  if [[ ${TIMEOUT} -gt 0 ]]; then
    let TIMEOUT-=1
    sleep 1
  else
    echo "ERROR: libvirt did not start in time (socket missing)"
    exit 1
  fi
done

# Rejoin libvirtd
wait
