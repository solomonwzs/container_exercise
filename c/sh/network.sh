#!/bin/bash
set -euo pipefail

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
bridge_addr="172.20.20.1/16"

# create network namespace
ip netns add "$network_ns"

# create a pair of veths
ip link add "$veth_A" type veth peer name "$veth_B"
ip link set "$veth_B" netns "$network_ns"

# create bridge
ip link add name "$bridge_name" type bridge
ip addr add "$bridge_addr" dev "$bridge_name"
ip link set "$bridge_name" up

# add veth to bridge
ip link set "$veth_A" master "$bridge_name"
ip link set "$veth_A" up

# set network in namespace
ip netns exec "$network_ns" ip link set lo up

function clean() {
    # clean
    ip link delete "$veth_A"
    ip link delete "$bridge_name" type bridge
    ip netns delete "$network_ns"
}

# exec test in namespace
echo "==="
ip netns exec "$network_ns" ip addr
echo "==="
ip netns exec "$network_ns" ping -c 1 127.0.0.1 || 1
echo "==="
ip netns exec "$network_ns" ping -c 1 baidu.com || 1
