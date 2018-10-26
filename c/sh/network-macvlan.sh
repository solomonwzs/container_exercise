#!/bin/bash
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-26
# @version  1.0
# @license  MIT

set -euo pipefail

EXECUTE_FILENAME=$(readlink -f "$0")
EXECUTE_DIRNAME=$(dirname "$EXECUTE_FILENAME")

source "${EXECUTE_DIRNAME}/common.sh"

while getopts "n:p:c:i:m:a:g:" opt; do
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
        m)
            mode=${OPTARG}
            ;;
        a)
            vname=${OPTARG}
            ;;
        g)
            gateway_ip=${OPTARG}
            ;;
        *)
            ;;
    esac
done

vlan_addr="${vlan_ip}/24"

mark="255.255.255.0"
mark_dec=$(ipv42dec "$mark")
vlan_ip_dec=$(ipv42dec "$vlan_ip")

rule_src="$(dec2ipv4 $(( vlan_ip_dec & mark_dec)))/24"

function delete() {
    iptables \
        -t nat \
        -D POSTROUTING \
        -s "$rule_src" \
        -j MASQUERADE || true

    ip netns exec "$network_ns" \
        ip link delete "$vname" || true
    ip netns delete "$network_ns" || true
    exit
}

function create() {
    # create network namespace
    ip netns add "$network_ns"

    # create vlan
    ip link add link "$interface" name "$vname" type macvlan mode "$mode"
    ip link set "$vname" netns "$network_ns"

    # set network in namespace
    ip netns exec "$network_ns" \
        ip link set lo up
    ip netns exec "$network_ns" \
        ip addr add "$vlan_addr" brd + dev "$vname"
    ip netns exec "$network_ns" \
        ip link set "$vname" up

    ip netns exec "$network_ns" \
        ip route add default via "$gateway_ip"

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
