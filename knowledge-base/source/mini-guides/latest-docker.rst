Using the latest Docker version
===============================

The virtual machines (VM) Harrow uses for running your tasks are based
on Ubuntu.  Even though Ubuntu offers Docker as part of its software
repositories, it is usually lagging some versions behind the latest
Docker version.  Sometimes you'd like to use the latest Docker version
to get the latest features.

Luckily, upgrading Docker is pretty easy, as it just a single file
that needs to be downloaded and put into the right place.  The actual
update is a three step process: stop the current docker daemon,
replace the docker binary with the new version, and then start the
docker daemon again.

The following snippet replaces Docker on the Harrow VM with the latest
available version.  Just place it at the top of your task to get the
lastest version of Docker.

.. code-block:: bash

   hfold update-docker
   PATH_TO_DOCKER=$(command -v docker)

   curl https://get.docker.com/builds/Linux/x86_64/docker-latest > docker
   chmod +x docker

   sudo systemctl stop docker

   sudo mv -v docker $PATH_TO_DOCKER

   sudo systemctl start docker

   hfold --end
