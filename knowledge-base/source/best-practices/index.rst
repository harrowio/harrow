Best Practices
==============

.. Complex Tasks
.. -------------
..
.. Above a certain level of complexity your Tasks_ should be moved into your Git
.. repository.
..
.. Above a few lines, the changes to the script become difficult to track, and
.. Harrow offers no versioning of task definitions.
..
.. It's entirely reasonable to have a separate repository with your build scripts
.. for each of your projects, or simply to have a branch in a specific repository
.. with your build scripts.
..
.. For example:
..
.. .. code-block:: bash
..
..   # Locally (your-project.git)
..   $ cd ./my-project/
..   $ git checkout --orphan harrow-tasks
..   $ vi deployment.sh
..   ... script body goes here ...
..
.. Then, in your task at Harrow, prefer to write something like this:
..
.. .. code-block:: bash
..
..   # In your Harrow task
..   $ git clone -b harrow-tasks ~/repositories/your-project ~/repositories/harrow-tasks
..   $ ./repositories/harrow-tasks/deployment.sh

Reusing Tasks
-------------

Tasks should make liberal use of Environment variables where appropriate. An
ideal example might be a simple task which includes the command ``cap
yourstage`` deploy:

.. code-block:: bash

  # In your Harrow task
  #!/bin/bash -e
  (
    cd repositories/your-project
    bundle install
    cap "$CAPSTAGE" deploy
  )

By not hard-coding the Capistrano stage, we can reuse this task to deploy
multiple environments ("stages" in Capistrano parlance).

For more complex scripts, including the case that the task definition resides
in a repository rather than directly in the Harrow task body field, it might be
worth while to add some guards around the use of the environment variable.

Capistrano, as it happens fails if the stage is not set, but incase an
unsuspecting user triggers this task in an unsupported environment, one lacking
the `CAPSTAGE` variable it would complain. Other tools may not.

Add a guard around important variables as follows:

.. code-block:: bash

  # In your Harrow task
  #!/bin/bash -e
  (
    cd repositories/your-project
    bundle install
    if [ -z "$CAPSTAGE" ]; then
      printf $(tput setaf 1) CAPSTAGE variable was not present in the ENV$(tput sgr0) 1>&2
      exit 123
    fi
    cap "$CAPSTAGE" deploy
  )

Environments
------------

Harrow creates a *Default* Environment in every new Project when they are
first created. You may remove it if you do not need it but it has been found
useful for the miscellanious tasks (such as testing and debugging new tasks).

Most people choose to separate their environments so that they have at least
"Production" and "Staging" environments, often tasks can be reused, as for most
people staging, and production environments tend to be similar.

Environments contain SSH keys, simple key-value pairs (Variables) and Secrets.
Secrets are not dissimilar to variables, but unprivilidged users (guest, and
member) cannot view the *value* of the secret, only it's exported name, this is
ideal for things like GPG keys, API tokens, etc.

For more information see Environments_.

Stopping Tasks on errors
------------------------

If you use Bash to run your tasks, starting the task body with the Bash shebang line:

.. code-block:: bash

  #!/bin/bash

It is important to note that in case of error, the script will not exit. This
means if your build relies on, for example uploading a file built earlier in
the process, without the Bash ``errexit`` option set, the script will not stop
if one if it's commands fails.

You can set the Bash ``errexit`` option in one of two ways:

.. code-block:: bash

  #!/bin/bash -e

.. code-block:: bash

  #!/bin/bash
  set -o errexit
  set -e         # same as `-o errexit`

Bash has a large number of `other useful options`_ which may also be of interest.

Avoiding Bash
-------------

Many people a dissuaded from shell scripting because Bash has a number of
pitfalls, if you count yourself in this number, you may write Harrow task
definitions in the language of your choice, at the slight mercy of which
versions of various other interpreters our host machines have installed.

To show a simple build script, in Bash and repeated in Ruby to serve as an
example below, something similar should work just as well with Python or Perl.

A more robust solution is to keep the task definition as short as possible,
host provisioning before deferring to scripts hosted in your own repositories.

.. code-block:: bash

  #!/bin/bash -e
  (
    cd repositories/my-repo
    npm install -g grunt
    grunt build
  )

This contrived example shows a simple Node.js repository being built with
Grunt, we could just as easily have written it with Ruby:

.. code-block:: bash

  #!/usr/bin/bin/env ruby -w

  Dir.chdir 'repositories/my-repo' do
    system 'npm install -g grunt'
    system 'grunt build'
  end

or, even with Perl:

.. code-block:: bash

  #!/usr/bin/bin/env perl

  chdir 'repositories/my-repo';
  system('npm install -g grunt');
  system('grunt build');

.. _Environments: ../environments/index.html
.. _Project: ../glossary/index.html#project
.. _Tasks: ../glossary/index.html#task
.. _other useful options: http://tldp.org/LDP/abs/html/options.html


If SSH authentication fails (cloning via HTTPS URLs)
-----------------------------------------------------

There is a number of configurations where the authentication via SSH keys can fail.
for example:
- adding 2 or more repositories from the same Git host using the SSH protocol in a Harrow project
- using tools within Harrow, such as Composer (PHP) or Bundler (Ruby), or even Capistrano (Ruby) when cloning private repositories over GIT+SSH.

In these cases, it's recommend to clone your repositories using HTTPS URLs with personal access tokens, instead of SSH keys. A token can be created as shown on `this Github guide`_
and the repository must be added in the format: https://username:password@host/<user>/<repo>.git,
or incase there is no username/password Github can use https://x:<token>@host/<user>/<repo>.git,
or you can drop the x: all together (then the Token is taken as a username) https://<token>@host/<user>/<repo>.git

Background: In a basic SSH repository authentication, SSH will connect using the first key it finds, if this key doesn't match the repository, the SSH backends from the Github/Bitbucket side will fail to authorize it.
It's a case of authentication vs authorization.
In a typical project where there are i.e three keys in the SSH agent: oneSSH deploy key, and one key each for two repositories, then one (more or less at random) will be selected by the SSH subsystem, and offered to Github.
Now, two of the three keys will "authenticate" you at Github (they are known to GitHub, known to belong to your user), but only one is “authorized” for a given repository, if you offer the wrong one, Github will reject your request.

If a repository is cloned using HTTPS URLs, the credential (token) becomes part of the URL, and is not selected at random from an agent with multiple cached credentials.

Furthermore, the Https protocol at Github is faster for cloning, as they can send larger packfiles.

.. figure:: ../quickstart/httpstoken.png


.. _this Github guide: https://help.github.com/articles/creating-an-access-token-for-command-line-use



How to enable access for Harrow on your firewall
------------------------------------------------

All the Harrow traffic passes through one of two gateways, The IP addresses to
be whitelisted on your firewall are:

.. code-block:: none

  144.76.166.35
  144.76.41.245
  144.76.62.172
  178.63.25.5
  52.19.120.94

You could verify which gateway any given operation uses by adding a job in
Harrow that gets your IP from dyndns. Simply create a task in Harrow with this
command:

.. code-block:: bash

  curl -s checkip.dyndns.org | sed -e 's/.*Current IP Address: //' -e 's/<.*$//'

Or shorter:

.. code-block:: bash

   wget http://ipinfo.io/ip -qO -

.. figure:: iptask.png

.. figure:: ipwhitelist.png
