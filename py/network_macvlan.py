#!/usr/bin/python3
# -*- coding: utf-8 -*-
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-30
# @version  1.0
# @license  MIT

from common_util import commands
import argparse
import json
import logging


logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)-15s [%(levelname)s] [%(filename)s:%(lineno)d]%(message)s'
)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Using network namespaces and a virtual switch to isolate servers")
    parser.add_argument("file", type=str, help="configure file")
    parser.add_argument("-c", "--command", type=str, help="Command")
    argv = parser.parse_args()

    with open(argv.file, "r") as fd:
        s = fd.read()
        conf = json.loads(s)
