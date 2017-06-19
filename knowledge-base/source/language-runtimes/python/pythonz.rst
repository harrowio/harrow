Using Python in Harrow via ``pythonz``
======================================

Using the pythonz_ offers the widest range of Python interpreters quickly
installed from binary packages where possible.

.. code-block:: bash

  #!/bin/bash -e

  sudo-pythonz install 2.7.3
  echo 'print "This line will be printed."' | tee example-script.py
  python --version
  python example-script.py

It is entirely possible to use Easy Install, PIP, VirtualEnv, etc simply
install them as you would normally.

.. _pythonz: https://github.com/saghul/pythonz
