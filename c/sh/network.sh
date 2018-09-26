#!/bin/bash
set -euo pipefail

BRIDGE_NAME="my_br"

ip link add name "$BRIDGE_NAME" type bridge
ip link set dev "$BRIDGE_NAME" up
