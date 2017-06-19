Interfacing with AWS
====================

With Harrow's features for scheduling and triggering jobs,
infrastructure can be managed automatically in a place accessible to
every team member.  This mini-guide details how to talk to AWS without
exposing your AWS credentials in an exploitable location.

Installing aws-cli in your task
-------------------------------

In order to use AWS' APIs, an API client is required.  The easiest way
is to install the official API client provided by Amazon.  You can add
it to your task with this short snippet of code:

.. code-block:: bash

  hfold install-awscli
  if ! which aws >/dev/null; then
    sudo apt-get -y install awscli
  fi
  hfold --end

Providing API credentials
-------------------------

The official Amazon API client can pick up API credentials from the
environment.  To securely add these credentials, edit the environment
in which you want to run your task and add two new secrets:
:code:`AWS_ACCESS_KEY_ID` and :code:`AWS_SECRET_ACCESS_KEY`.  The values for these
secrets are your access credentials to Amazon's API.

Once you have added those secrets, the `aws` command in your task will
automatically pick them up and be able to connect to the AWS API.
