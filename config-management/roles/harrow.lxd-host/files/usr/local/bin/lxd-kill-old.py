#!/usr/bin/env python3

import sys
import subprocess
import datetime
import dateutil.parser
import json
import time
from collections import namedtuple

Container = namedtuple('Container', ('name', 'created_at'))


def main(**kwargs):
    dispatcher = {
        'list': fn_list,
        'kill': fn_kill,
        'help': lambda: print(
            "Usage: {} [help|list|kill]\n".format(
                sys.argv[0])),
    }
    fn = dispatcher.get(
        kwargs["action"],
        lambda: print("no match for `{}'\n".format(
            kwargs['action'])))
    fn()


def fn_kill():
    for container in containers():
        try:
            out = subprocess.check_output(
                ["lxc", "delete", "--force", container.name])
            print("> deleted container {container}\n\t{out}".format(
                container=container, out=out.decode("utf-8")))
        except subprocess.CalledProcessError as e:
            print(
                "! couldn't delete container {container}".format(
                    container=container))


def fn_list():
    for container in containers():
        print(
            "\t{name} ({created_at})\n".format(
                name=container.name,
                created_at=container.created_at))


def containers():
    try:
        lst = subprocess.check_output(["lxc", "list", "--format=json"])
        jsn = json.loads(lst.decode("utf-8"))
        containers = map(
            lambda c: Container(
                c["name"],
                dateutil.parser.parse(
                    c["created_at"])),
            jsn)
        return filter(
            lambda c: c.created_at < datetime.datetime.now(
                datetime.timezone.utc) -
            datetime.timedelta(
                hours=2),
            containers)
    except subprocess.CalledProcessError as e:
        return []

main(action=sys.argv[1])
