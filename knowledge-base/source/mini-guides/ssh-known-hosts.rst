SSH Known Hosts Mini-Guide
==========================

The SSH known hosts file stored at ``/etc/ssh/known_hosts`` or more commonly at
``~/.ssh/known_hosts`` contains a list of host names or IP addresses and their
corresponding *host fingerprint*.

This mechanism is designed to protect you against man in the middle attacks, by
forwarding keys, or uploading code or assets to a server that may have been
replaced with a compromised one.

Harrow takes security very seriously, and where other continuous integration
services might default to accepting untrusted unknown hosts, Harrow tries to
help you maintain a secure environment.

If you see a warning about SSH Known Hosts in your log output, you can solve it
in one of two ways:

The Secure Way
--------------

With many servers, the best option is to commit an SSH known hosts file to your
repository, and to move into the path where SSH expects to find it at the start
of your build:

.. code-block:: bash

  cp repositories/my-project/.ssh/known_hosts ~/.ssh/known_hosts

This way even outside Harrow your developers don't have to risk trusting
untrusted servers as the ``known_hosts`` list can be shared amongst your
developers, likely the list of deploy servers resides in the repository anyway,
so a little dilligence around server identity security isn't a huge leap.

As an easier solution, if this project only targets one, or a small number of
servers it's possible to add an environment secret to the project, with the SSH
host key, and simply write it out into the file:

  echo "my-server-address.com\t$MY_SERVER_HOST_KEY_ENV_SECRET" >> ~/.ssh/known_hosts

Then simply make sure there's an environment secret with the name
``MY_SERVER_HOST_KEY_ENV_SECRET`` in the Project environment.

The Quick Way
-------------

The quick way includes using ``ssh-keyscan`` to fetch the key from the server,
and write it directly into the ``~/.ssh/known_hosts/`` file.

.. code-block:: bash

  ssh-keyscan -4 my-server-address.com > ~/.ssh/known_hosts

This carries some risks, but for most applications most of the time should be
mostly secure. The attack vector is that if someone is able to spoof the DNS or
claim your IP address, you would offer your private keys or passwords to a
server that you did not trust, and had not seen before.

This is the way we document in our quick-start guides, whilst slightly less
secure, it's much easier to get started, and many deployment tools are not as
strict as we are with known host security anyway.
