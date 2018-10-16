#!/bin/bash
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-15
# @version  1.0
# @license  MIT

set -euo pipefail

EXECUTE_FILENAME=$(readlink -f "$0")
EXECUTE_DIRNAME=$(dirname "$EXECUTE_FILENAME")

source "${EXECUTE_DIRNAME}/common.sh"

while getopts "n:i:c:" opt; do
    case "${opt}" in
        n)
            network_ns=${OPTARG}
            ;;
        i)
            veth_ip=${OPTARG}
            ;;
        c)
            cmd=${OPTARG}
            ;;
        *)
            ;;
    esac
done

veth_A="${network_ns}-veth-A"
veth_B="${network_ns}-veth-B"
bridge_name="${network_ns}-bridge"

mark="255.255.255.0"
mark_dec=$(ipv42dec "$mark")
veth_ip_dec=$(ipv42dec "$veth_ip")

route_ip=$(dec2ipv4 $(( veth_ip_dec & mark_dec | 1 )))
bridge_addr="${route_ip}/24"
veth_addr="${veth_ip}/24"

rule_src="$(dec2ipv4 $(( veth_ip_dec & mark_dec)))/24"

function delete() {
    # sysctl -w net.ipv4.ip_forward=0
    iptables \
        -t nat \
        -D POSTROUTING \
        -s "$rule_src" \
        -j MASQUERADE || true

    ip link delete "$veth_A" || true
    ip link delete "$bridge_name" type bridge || true
    ip netns delete "$network_ns" || true
    exit
}

function create() {
    # create network namespace
    ip netns add "$network_ns"

    # create a pair of veths
    ip link add "$veth_A" type veth peer name "$veth_B"
    ip link set "$veth_B" netns "$network_ns"

    # create bridge
    ip link add name "$bridge_name" type bridge
    ip addr add "$bridge_addr" brd + dev "$bridge_name"
    ip link set "$bridge_name" up

    # add veth to bridge
    ip link set "$veth_A" master "$bridge_name"
    ip link set "$veth_A" up

    # set network in namespace
    ip netns exec "$network_ns" \
        ip link set lo up
    ip netns exec "$network_ns" \
        ip link set "$veth_B" name eth0
    ip netns exec "$network_ns" \
        ip link set eth0 up
    ip netns exec "$network_ns" \
        ip addr add "$veth_addr" dev eth0
    ip netns exec "$network_ns" \
        ip route add default via "$route_ip"

    # enable to reach the internet
    iptables \
        -t nat \
        -A POSTROUTING \
        -s "$rule_src" \
        -j MASQUERADE
    sysctl -w net.ipv4.ip_forward=1
}

function test_ns() {
    # exec test in namespace
    pmsg "===="
    ip route show

    pmsg "===="
    route -n

    pmsg "===="
    ip netns exec "$network_ns" \
        route -n

    pmsg "===="
    ip addr

    pmsg "===="
    ip netns exec "$network_ns" \
        ip addr

    pmsg "===="
    ip netns exec "$network_ns" \
        ping -c 1 -W 1 127.0.0.1

    pmsg "===="
    ip netns exec "$network_ns" \
        ping -c 1 -W 1 "$veth_ip"

    pmsg "==="
    ip netns exec "$network_ns" \
        ping -c 1 -W 1 220.181.57.216

    pmsg "==="
    ip netns exec "$network_ns" \
        ping -c 1 baidu.com || clean
}

# trap "clean" SIGINT SIGTERM

case "$cmd" in
    new)
        create
        ;;
    del)
        delete
        ;;
    test)
        test_ns
        ;;
    *)
        perr "unknown command: ${cmd}"
        ;;
esac
