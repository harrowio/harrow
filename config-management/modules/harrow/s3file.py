#!/usr/bin/env python
# WANT_JSON

import json
import sys
import os
import subprocess
import shutil
import tempfile
import grp
import pwd

from ansible.module_utils.basic import AnsibleModule

def main():
    module = AnsibleModule(
        argument_spec = dict(
            bucket            = dict(required = True),
            key               = dict(required = True),
            dest              = dict(required = True),
            mode              = dict(required = True, type = "raw"),
            owner             = dict(required = True),
            group             = dict(required = True),
            access_key_id     = dict(required = True),
            secret_access_key = dict(required = True),
            region            = dict(required = True)
        )
    )

    os.environ["AWS_ACCESS_KEY_ID"] = module.params["access_key_id"]
    os.environ["AWS_SECRET_ACCESS_KEY"] = module.params["secret_access_key"]

    head_object = subprocess.check_output(["/usr/bin/aws", "s3api", "head-object",
        "--region", module.params["region"],
        "--bucket", module.params["bucket"],
        "--key", module.params["key"]])

    res=json.loads(head_object)
    etag=res["ETag"]

    if not os.path.exists("/var/local"):
        os.mkdir("/var/local")
    if not os.path.exists("/var/local/s3files"):
        os.mkdir("/var/local/s3files", 0700)

    etag_name = "/var/local/s3files/{0}.etag".format(module.params["key"].replace("/","___"))
    name = "/var/local/s3files/{0}".format(module.params["key"].replace("/","___"))

    if os.path.exists(etag_name) and etag == read_etag(etag_name):
        module.exit_json(changed=False)

    subprocess.check_call(["/usr/bin/aws", "s3", "cp",
        "s3://{0}/{1}".format(module.params["bucket"], module.params["key"]),
        name,
        "--region", module.params["region"]])

    tmpdir = tempfile.mkdtemp()
    tmp = "{0}/tmp".format(tmpdir)
    shutil.copy(name, tmp)

    os.chmod(tmp, module.params["mode"])

    uid = pwd.getpwnam(module.params["owner"]).pw_uid
    gid = grp.getgrnam(module.params["group"]).gr_gid
    os.chown(tmp, uid, gid)
    shutil.move(tmp, module.params["dest"])

    os.rmdir(tmpdir)
    write_etag(etag_name, etag)
    module.exit_json(changed=True)

def read_etag(name):
    with open(name, "r") as f:
        return f.read()

def write_etag(name, etag):
    with open(name, "wt") as f:
        os.chmod(name, 0600)
        f.write(etag)

if __name__ == '__main__':
    main()
