#!/bin/bash
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-11
# @version  1.0
# @license  MIT

set -euo pipefail

SCRIPT=$(readlink -f "$0")
DIR=$(dirname "$SCRIPT")

BASESYS_PATH=/home/solomon/workspace/basesystem/ubuntu_xenial_1604/

"${DIR}/../build/container-test" \
    -h "container" \
    -p "${BASESYS_PATH}" \
    -m "${DIR}/network.sh:/usr/local/bin/network.sh" \
    -e "PATH=/usr/bin:/usr/local/bin:/usr/local/sbin:/bin:/sbin"
