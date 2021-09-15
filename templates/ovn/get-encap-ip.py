#!/usr/bin/env python3
import os

import netaddr
import netifaces


def get_overlay_cidrs():
    val = os.environ.get('OVERLAY_CIDRS')
    if not val:
        return []

    return [netaddr.IPNetwork(c) for c in val.split(',')]


def is_overlay_ip(ipnet, overlay_cidrs):
    for overlay_cidr in overlay_cidrs:
        if ipnet in overlay_cidr:
            return True

    return False


def match_encap_ip(overlay_cidrs):
    for iface in netifaces.interfaces():
        ifaddrs = netifaces.ifaddresses(iface)
        for ifaddr in ifaddrs.get(netifaces.AF_INET, []):
            ipnet = netaddr.IPNetwork(ifaddr['addr'] + '/' + ifaddr['netmask'])
            if is_overlay_ip(ipnet, overlay_cidrs):
                return ifaddr['addr']

    raise ValueError('No interfaces matched OVERLAY_CIDRS', overlay_cidrs)


def default_iface_ip():
    gateways = netifaces.gateways()
    _, gw_iface = gateways['default'][netifaces.AF_INET]

    ifaddrs = netifaces.ifaddresses(gw_iface)
    ifaddr = ifaddrs[netifaces.AF_INET][0]
    return ifaddr['addr']


def guess_encap_ip():
    overlay_cidrs = get_overlay_cidrs()
    if overlay_cidrs:
        return match_encap_ip(overlay_cidrs)

    return default_iface_ip()


def main():
    print(guess_encap_ip())

main()
