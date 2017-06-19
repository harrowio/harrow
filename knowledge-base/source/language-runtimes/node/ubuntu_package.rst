Using Node.js in Harrow via the official NodeSource Ubuntu Package
==================================================================

Using the `official Node.js package`_ is an easy and reliable ways to use
Node.js on Harrow.

.. code-block:: bash

  #!/bin/bash -e

  hfold "Installing Node.js via NodeSource.com package"
  curl --silent --location https://deb.nodesource.com/setup_0.12 | sudo bash -
  sudo apt-get install --yes nodejs
  hfold --end

Note you might want to install the ``build-essential`` along side ``nodejs`` if you
have NPM packages which might be backed by C libraries, or similar.

.. code-block:: bash

  #!/bin/bash -e

  hfold "Installing Node.js via NodeSource.com package"
  curl --silent --location https://deb.nodesource.com/setup_0.12 | sudo bash -
  sudo apt-get install --yes nodejs
  hfold --end

  echo 'console.log("Hello World");' | tee > example-script.js
  node example-script.js

.. _official Node.js package: https://github.com/nodejs/node-v0.x-archive/wiki/Installing-Node.js-via-package-manager#debian-and-ubuntu-based-linux-distributions
