Environments
============

Harrow defines a Project as having multiple *Environments* which properties
such as variables, deploy keys, and secrets.

Environments are intended to allow `Tasks` such as "Deploy the website" or "run
the acceptance test suite" to be reused across different environments,
inheriting their configuration, or some of their configuration from the
Environment.

.. toctree::
   :maxdepth: 2

   Variables <variables/index>
   SSH (Deploy) Keys <ssh-keys/index>
   Secrets <secrets/index>
