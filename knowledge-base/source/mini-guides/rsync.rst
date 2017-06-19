Deploy files to Server with Rsync
=================================

| In Devops, it's often needed to copy a set of files from a local repository to a server.

| Among other methods, you can use `rsync`_ for this purpose.
|
| In this guide we will create a Harrow task that copies your files to the desired serve and we will exclude the `.git` directory, to make sure that the files that contain private data and are not necessary for the app view are not copied over and exposed.
|
|

.. code-block:: bash

  #!/bin/bash -e
  ssh-keyscan -4 <server_IP_address> ~/.ssh/known_hosts
  (
    cd repositories/<repository_name>
    rsync -a --exclude='.git/' ./* <ssh_user>@<server_IP_address>:/path/on/server/
  )


It is a good practice, if you want to make this task reusable, to set $SSH_USER , $IP_ADDRESS and $PATH_ON_SERVER as `variables`_, and move them to an environment.

.. _rsync: http://linux.die.net/man/1/rsync
.. _variables: https://help.ubuntu.com/community/EnvironmentVariables
