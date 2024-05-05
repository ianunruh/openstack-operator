#!/usr/bin/env python3
import ipaddress
import os
import subprocess


def get_overlay_cidrs() -> list[ipaddress.IPv4Network]:
    val = os.environ.get('OVERLAY_CIDRS')
    if not val:
        return []

    return [ipaddress.ip_network(c) for c in val.split(',')]


def is_overlay_ip(addr: ipaddress.IPv4Address, overlay_cidrs: list[ipaddress.IPv4Network]) -> bool:
    for overlay_cidr in overlay_cidrs:
        if addr in overlay_cidr:
            return True

    return False


def run_cmd(cmd) -> list[str]:
    output = subprocess.check_output(['bash', '-c', cmd]).decode('utf-8')
    return [line.strip() for line in output.strip().splitlines()]


def get_host_interfaces() -> list[ipaddress.IPv4Interface]:
    # inet 127.0.0.1/8 scope host lo
    # inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
    return [ipaddress.IPv4Interface(line.split(' ')[1])
            for line in run_cmd('ip -4 addr show | grep inet')]


def match_encap_ip(overlay_cidrs: list[ipaddress.IPv4Network]) -> ipaddress.IPv4Address:
    for iface in get_host_interfaces():
        if is_overlay_ip(iface, overlay_cidrs):
            return iface.ip

    raise ValueError('No interfaces matched OVERLAY_CIDRS', overlay_cidrs)


def default_iface_ip() -> ipaddress.IPv4Network:
    # default via 172.17.0.1 dev eth0
    for route_line in run_cmd('ip -4 route show default'):
        iface_name = route_line.split(' ')[4]
        # inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
        for line in run_cmd(f'ip -4 addr show dev {iface_name} | grep inet'):
            return ipaddress.IPv4Interface(line.split(' ')[1]).ip


def get_encap_ip() -> ipaddress.IPv4Network:
    overlay_cidrs = get_overlay_cidrs()
    if overlay_cidrs:
        return match_encap_ip(overlay_cidrs)

    return default_iface_ip()


def main():
    print(get_encap_ip())

main()
