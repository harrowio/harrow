Using Elasticsearch In Harrow With Docker
=========================================

The `official Elasticsearch Docker image`_ can easily be fetched into the
Harrow environment with the following command. Beware the image will not
persist any data (for information about persisting data see caching_).

.. code-block:: bash

  $ hfold "Starting Elasticsearch"
  $ sudo docker run -d -p 9200:9200 -p 9300:9300 elasticsearch
  $ hfold --end

.. note:: To learn more about the ``hfold`` command see `Harrow Utilities -> hfold`_.

With the ElasticSearch daemon running it's easily connected to using ``curl``
(for more examples see the `Elasticsearch documentation`_):

.. code-block:: bash

  $ curl -q 127.0.0.1:9200
  {
    "status" : 200,
    "name" : "Living Laser",
    "cluster_name" : "elasticsearch",
    "version" : {
      "number" : "1.7.2",
      "build_hash" : "e43676b1385b8125d647f593f7202acbd816e8ec",
      "build_timestamp" : "2015-09-14T09:49:53Z",
      "build_snapshot" : false,
      "lucene_version" : "4.10.4"
    },
    "tagline" : "You Know, for Search"
  }

Be aware that ElasticSearch is installed in a Docker container, and as such the
client libraries and other parts of the package and it's dependencies may not
be available in the operation host machine.

If your application needs to use the Elasticsearch client, please make sure
your application specifies it's own dependencies (e.g Ruby Gem), which are
installed at runtime.

.. _Harrow Utilities -> hfold: ../../harrow-utilities/hfold
.. _caching: ../../../caching/index
.. _official Elasticsearch Docker image: https://hub.docker.com/_/elasticsearch/
.. _Elasticsearch documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html
