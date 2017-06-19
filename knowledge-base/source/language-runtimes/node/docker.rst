Using Node.js in Harrow via the official Docker container
=========================================================

Using the `official Node.js Docker containers`_ is one of the fastest and most
reliable ways to use Node.js on Harrow.

The containers are pulled directly each time, which can be time consuming (a
few seconds), but will eventually be cached in a Harrow-local Docker repository
mirror.

.. code-block:: bash

  #!/bin/bash -e

  echo 'console.log("Hello World");' | tee > example-script.js
  sudo docker run -v "$PWD":/usr/src/example -w /usr/src/example node:latest node example-script.js

.. _official Node.js Docker containers: https://hub.docker.com/_/node/
