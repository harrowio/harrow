Docker
======

The `official Elasticsearch Docker image`_ can easily be fetched into the
Harrow environment with the following command. Beware the image will not
persist any data (for information about persisting data see caching_

.. todo::
  Write This Section

.. code-block:: bash

  $ hfold "Starting Elasticsearch"
  $ docker run -d elasticsearch elasticsearch -Des.node.name="TestNode"
  $ hfold --end

.. note:: To learn more about the `hfold` command see hfold_.

.. _hfold: /harrow-utilities/hfold
.. _caching: /caching/index
.. _official Elasticsearch Docker image: https://hub.docker.com/_/elasticsearch/
