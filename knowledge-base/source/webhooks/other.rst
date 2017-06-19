Other
=====

Webhooks can be triggered with any method (``GET``, ``POST``, ``PUT``,
``DELETE``), any request sent to the URL generated will trigger your operation.

Your operation can examine the body passed to it via the webhook's ``POST``
body by looking into the file:

.. code-block:: bash

  $ cat ./harrow-webhook-body

.. note::
  This file will always have the suffix ``.json``, even if the webhook body is
  not JSON. This may be subject to change with advanced warning.

Triggering webhooks with ``curl``
---------------------------------

If you don't have a specific integration in mind (Github, Bitbucket, etc) you
can easily trigger a webhook as follows:

.. code-block:: bash

  $ curl -X POST -d @something.json https://www.app.harrow.io/api/wh/82412f59b547d123

This will trigger the job referenced by the webhook, and post a local file
``something.json`` in the body.

You can examine the contents of the posted file using something like the
following in a Harrow task:

.. code-block:: bash

  #!/bin/bash -e

  cat harrow-webhook-body
  echo
  printf "$(tput setaf 2)Finished OK\n$(tput sgr0)"

Triggering Webhooks on selected branch changes
---------------------------------------------------

By default, webhooks trigger on every branch change on GitHub/Bitbucket. 
If you need to trigger a Harrow task only when a specific branch change is pushed,  you can add the following check to your task script. 

.. code-block:: bash

    (
    	cd ~/repositories/$GIT_REPOSITORY
    	if [ -e ~/harrow-webhook-body ] && ! [ $(git symbolic-ref --short HEAD) = $GIT_BRANCH ]; then 
        printf "Exiting, wanted branch $GIT_BRANCH got $(git symbolic-ref --short HEAD)" 1>&2 
        exit 0; 
    	fi
    )	

The file ~/harrow-webhook-body only exists if the Harrow job is triggered by a webhook. With this check it can be run by hand and by webhook only on the selected branch

NB: $GIT_REPOSITORY and $GIT_BRANCH should be added as environment variables to your environment. 


Tasks may not always be triggered by webhook!
---------------------------------------------

Be aware that the ``harrow-webhook-body`` file only exists in case that the
operation was triggered by a webhook, it is easy to check if the file is
present and change behaviour accordingly:

.. code-block:: bash

  #!/bin/bash -e

  if [ ! -f harrow-webhook-body]; then
    printf $(tput setaf 1) harrow-webhook-body was not present. I was not triggered via webhook.$(tput sgr0) 1>&2
  fi
  ... continue as normal, if we need that file's content ...

