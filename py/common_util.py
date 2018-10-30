#!/usr/bin/python3
# -*- coding: utf-8 -*-
#
# @author   Solomon Ng <solomon.wzs@gmail.com>
# @date     2018-10-30
# @version  1.0
# @license  MIT


from subprocess import PIPE
from subprocess import Popen
import logging


def commands(cmds):
    for cmd in cmds:
        logging.debug(f"\033[1;32m{cmd}\033[0m")
        p = Popen(cmd, shell=True, stdout=PIPE, stderr=PIPE)
        output, err = p.communicate()
        if p.returncode != 0:
            logging.error(err)
            return False
        elif len(output) != 0:
            logging.debug(output)
    return True
