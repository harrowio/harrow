Using Python in Harrow via the official Docker container
=========================================================

Using the `official Python Docker containers`_ is one of the fastest and most
reliable ways to use Python on Harrow.

The containers are pulled directly each time, which can be time consuming (a
few seconds), but will eventually be cached in a Harrow-local Docker repository
mirror.

.. code-block:: bash

  #!/bin/bash -e

  echo 'print "This line will be printed."' | tee example-script.py
  sudo docker run -v "$PWD":/usr/src/example -w /usr/src/example python:latest python example-script.py

This can be taken further to run daemons in Python in the background by
detaching the Docker container with ``-d``:

.. code-block:: bash

  #!/bin/bash -e

  sudo docker run -d -v "$PWD":/usr/src/example -w /usr/src/example python:latest python example-script.py

This doesn't make sense for this short lived example script, but for some
daemons it might make sense.

.. _official Python Docker containers: https://hub.docker.com/_/python/
