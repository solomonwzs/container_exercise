#!/bin/bash
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-11
# @version  1.0
# @license  MIT

set -euo pipefail

SCRIPT=$(readlink -f "$0")
DIR=$(dirname "$SCRIPT")
SRC_DIR="${DIR}/../"
BUILD_DIR="${DIR}/../build"

BASESYS_PATH=/home/solomon/workspace/basesystem/ubuntu_xenial_1604/

cd "${BUILD_DIR}"

cmake "$SRC_DIR"

make

"${BUILD_DIR}/container-test" \
    -h "container" \
    -p "${BASESYS_PATH}" \
    -m "${DIR}/network.sh:/usr/local/bin/network.sh" \
    -e "PATH=/usr/bin:/usr/local/bin:/usr/local/sbin:/bin:/sbin"
