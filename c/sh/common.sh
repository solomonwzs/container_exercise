#!/bin/bash
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-15
# @version  1.0
# @license  MIT

set -euo pipefail


_defer_mission=()


function pmsg() {
    echo -e "\\e[1;32m${1}\\e[0m"
}


function perr() {
    echo -e "\\e[1;31m${1}\\e[0m" && exit
}


function ipv42dec() {
    declare -i a b c d;
    IFS=. read -r a b c d <<<"$1";
    echo "$(((a << 24) + (b << 16) + (c << 8) + d))";
}


function dec2ipv4() {
    declare -i a=$((~(-1 << 8))) b=$1;
    set -- "$((b >> 24 & a))" "$((b >> 16 & a))" "$((b >> 8 & a))"\
        "$((b & a))";
    local IFS=.;
    echo "$*";
}


function defer_init() {
    _defer_mission=()
}


function defer_add() {
    _defer_mission+=("$1")
}


function defer_run() {
    for i in $(seq $((${#_defer_mission[@]} - 1)) -1 0); do
        ${_defer_mission[i]}
    done
    exit
}


if [ "$(readlink -f "$0")" == "$(readlink -f "${BASH_SOURCE[0]}")" ]; then
    ip="192.168.197.130"
    ip2="255.255.255.0"

    a=$(ipv42dec "$ip")
    b=$(dec2ipv4 "$a")

    c=$(ipv42dec "$ip2")
    d=$((a & c | 1))
    e=$(dec2ipv4 "$d")

    echo "$a"
    echo "$b"
    echo "$e"

    defer_init
    defer_add "echo 123"
    defer_add "echo 456"
    defer_add "echo 789"
    defer_run
fi
