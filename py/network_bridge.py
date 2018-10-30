#!/usr/bin/python3
# -*- coding: utf-8 -*-
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-16
# @version  1.0
# @license  MIT

from common_util import commands
from ipaddress import IPv4Address
import argparse
import json
import logging


logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)-15s [%(levelname)s] [%(filename)s:%(lineno)d]%(message)s'
)


def create_namespace_network(bridge_name, route_ip, network_ns, veth_ip):
    veth_A = f"{network_ns}-veth-A"
    veth_B = f"{network_ns}-veth-B"
    veth_addr = f"{veth_ip}/24"
    ns_exec = f"ip netns exec {network_ns}"

    # create network namespace
    commands([f"ip netns add {network_ns}"])

    # create a pair of veth
    commands([
        f"ip link add {veth_A} type veth peer name {veth_B}",
        f"ip link set {veth_B} netns {network_ns}",
    ])

    # set veth to bridge
    commands([
        f"ip link set {veth_A} master {bridge_name}",
        f"ip link set {veth_A} up",
    ])

    # set network in namespace
    commands([
        f"{ns_exec} ip link set lo up",
        f"{ns_exec} ip link set {veth_B} name eth0",
        f"{ns_exec} ip link set eth0 up",
        f"{ns_exec} ip addr add {veth_addr} dev eth0",
        f"{ns_exec} ip route add default via {route_ip}",
    ])


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Using network namespaces and a virtual switch to isolate servers")
    parser.add_argument("file", type=str, help="configure file")
    parser.add_argument("-c", "--command", type=str, help="Command")
    argv = parser.parse_args()

    with open(argv.file, "r") as fd:
        s = fd.read()
        conf = json.loads(s)

    bridge_name = conf["bridge"]
    route_ip = conf["route"]
    bridge_addr = f"{route_ip}/24"

    rule_src = str(IPv4Address(
        int(IPv4Address(route_ip)) & int(IPv4Address("255.255.255.0"))))
    rule_src = f"{rule_src}/24"

    if argv.command == "new":
        # create bridge
        commands([
            f"ip link add name {bridge_name} type bridge",
            f"ip addr add {bridge_addr} brd + dev {bridge_name}",
            f"ip link set {bridge_name} up",
        ])

        for network_ns, c in conf["namespaces"].items():
            veth_ip = c["ip"]
            create_namespace_network(bridge_name, route_ip, network_ns,
                                     veth_ip)

        commands([
            f"iptables -t nat -A POSTROUTING -s {rule_src} -j MASQUERADE",
            f"sysctl -w net.ipv4.ip_forward=1",
        ])
    elif argv.command == "del":
        commands([
            f"iptables -t nat -D POSTROUTING -s {rule_src} -j MASQUERADE",
        ])
        for network_ns in conf["namespaces"].keys():
            veth_A = f"{network_ns}-veth-A"
            commands([
                f"ip link delete {veth_A}",
                f"ip netns delete {network_ns}",
            ])
        commands([f"ip link delete {bridge_name}"])
    elif argv.command == "test":
        for network_ns, c in conf["namespaces"].items():
            veth_ip = c["ip"]
            commands([f"ping -c 1 {veth_ip}"])

            ns_exec = f"ip netns exec {network_ns}"
            for ns, c in conf["namespaces"].items():
                veth_ip = c["ip"]
                commands([f"{ns_exec} ping -c 1 {veth_ip}"])
        commands([f"{ns_exec} ping -c 1 baidu.com"])
