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

# while getopts "b:n:A:B:i:" opt; do
#     case "${opt}" in
#         b)
#             bridge_name=${OPTARG}
#             ;;
#         n)
#             network_ns=${OPTARG}
#             ;;
#         A)
#             veth_A=${OPTARG}
#             ;;
#         B)
#             veth_B=${OPTARG}
#             ;;
#         *)
#             ;;
#     esac
# done

network_ns="my-ns-1"
veth_A="my-veth-A"
veth_B="my-veth-B"
bridge_name="my-bridge-1"
route_ip="172.20.10.1"
bridge_addr="${route_ip}/24"
veth_addr="172.20.10.11/24"

function clean() {
    sysctl -w net.ipv4.ip_forward=0
    iptables \
        -t nat \
        -D POSTROUTING \
        -s 172.20.10.0/24 \
        -j MASQUERADE || true

    ip link delete "$veth_A" || true
    ip link delete "$bridge_name" type bridge || true
    ip netns delete "$network_ns" || true
    exit
}

trap "clean" SIGINT SIGTERM

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
    -s 172.20.10.0/24 \
    -j MASQUERADE
sysctl -w net.ipv4.ip_forward=1

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
    ping -c 1 -W 1 127.0.0.1 || clean
pmsg "===="
ip netns exec "$network_ns" \
    ping -c 1 -W 1 172.20.10.11 || clean
pmsg "==="
ip netns exec "$network_ns" \
    ping -c 1 -W 1 220.181.57.216 || clean
# ip netns exec "$network_ns" \
#     ping -c 1 baidu.com || clean

clean
