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

while getopts "n:p:c:i:d:" opt; do
    case "${opt}" in
        n)
            network_ns=${OPTARG}
            ;;
        p)
            vlan_ip=${OPTARG}
            ;;
        c)
            cmd=${OPTARG}
            ;;
        i)
            interface=${OPTARG}
            ;;
        d)
            vlan_id=${OPTARG}
            ;;
        *)
            ;;
    esac
done

vlan_name="${interface}.${vlan_id}"
vlan_addr="${vlan_ip}/24"

mask="255.255.255.0"
mask_dec=$(ipv42dec "$mask")
vlan_ip_dec=$(ipv42dec "$vlan_ip")

rule_src="$(dec2ipv4 $(( vlan_ip_dec & mask_dec)))/24"

function delete() {
    iptables \
        -t nat \
        -D POSTROUTING \
        -s "$rule_src" \
        -j MASQUERADE || true

    ip netns exec "$network_ns" \
        ip link delete "$vlan_name" || true
    ip netns delete "$network_ns" || true
    exit
}

function create() {
    # create network namespace
    ip netns add "$network_ns"

    # create vlan
    ip link add link "$interface" name "$vlan_name" type vlan id "$vlan_id"
    ip link set "$vlan_name" netns "$network_ns"

    # set network in namespace
    ip netns exec "$network_ns" \
        ip link set lo up
    ip netns exec "$network_ns" \
        ip addr add "$vlan_addr" brd + dev "$vlan_name"
    ip netns exec "$network_ns" \
        ip link set "$vlan_name" up

    iptables \
        -t nat \
        -A POSTROUTING \
        -s "$rule_src" \
        -j MASQUERADE || true
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
        ping -c 1 -W 1 "$vlan_ip"

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
