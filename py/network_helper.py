#!/usr/bin/python3
# -*- coding: utf-8 -*-
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-16
# @version  1.0
# @license  MIT

from subprocess import (
    Popen,
    PIPE,
)
from ipaddress import IPv4Address
import logging


logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)-15s [%(levelname)s] [%(filename)s:%(lineno)d]%(message)s'
)


def commands(cmds):
    for cmd in cmds:
        p = Popen(cmd, shell=True, stdout=PIPE, stderr=PIPE)
        output, err = p.communicate()
        if p.returncode != 0:
            logging.error(err)
            return False
        elif len(output) != 0:
            logging.debug(output)
    return True


def create_network_namespace(network_ns):
    return commands([f"ip netns add {network_ns}"])


def delete_network_namespace(network_ns):
    return commands([f"ip netns delete {network_ns}"])


def create_veths(network_ns):
    pass


if __name__ == "__main__":
    print(int(IPv4Address('127.0.0.1')))
