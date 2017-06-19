Native
======

All Harrow operations are run inside a virtual machine granted ``root`` access
to install and configure any software or required services.

The default stack is modern Ubuntu, a Debian derived distribution which is
widely supported.

Packages can simply be installed with Aptitude_.

Best Practices
--------------

Avoiding prompts:

.. code-block:: bash

  $ apt-get install -y my-package
                  # ^^ prevents aptitude trying to prompt for input

We have configured aptitude never to prompt for input, which would cause it to
stall, without positive encouragement that causes it to assume ``no`` would be
the answer to potentially risky operations, such as changing the set of
installed system packages.

Use A Custom Apt-Repository
---------------------------

You can simply add a custom apt repository or PPA (Private Package Archive)
using the following:

.. code-block:: bash

  $ sudo add-apt-repository -y ppa:webupd8team/java
  $ sudo apt-get update

The ``apt-get update`` is important because without it the packages are not
available.

It's important to run ``apt-get update`` regularly if you want to stay up to
date. We periodically update the base image so that the update doesn't always
have to fetch a tonne of new and updated packages taking time out of your
builds.

.. _Aptitude: https://help.ubuntu.com/lts/serverguide/aptitude.html
