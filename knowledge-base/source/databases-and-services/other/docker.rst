Docker
======

**Any** Docker image can be pulled and started inside your Harrow container.

.. code-block:: bash

  $ hfold "Starting Elasticsearch"
  $ sudo docker run -d redis:latest -p 6379:6379
  $ hfold --end

.. note:: To learn more about the `hfold` command see hfold_.

It is important to remember to:

#. Use ``sudo`` to run the Docker command line tool.
#. Remember to map ports with ``-p 1234:1234`` if you aren't linking multiple
   Containers.
#. Your data volumes will not be persisted between Operations unless you use
   `filesystem caching`_.

.. _hfold: /harrow-utilities/hfold
.. _filesystem caching: /caching/index
