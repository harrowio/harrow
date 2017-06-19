Using Elasticsearch On Harrow With A Native Linux Package
=========================================================

Installing Elasticsearch natively requires a modern Java VM and some updates to
to the ``apt`` package sources:

.. code-block:: bash

  $ hfold "Installing Elasticsearch (Native)"
  $ sudo add-apt-repository -y ppa:webupd8team/java
  $ sudo apt-get update
  $ sudo apt-get -y install oracle-java8-installer
  $ wget -O - http://packages.elasticsearch.org/GPG-KEY-elasticsearch | sudo apt-key add -
  $ echo 'deb http://packages.elasticsearch.org/elasticsearch/1.4/debian stable main' | sudo tee /etc/apt/sources.list.d/elasticsearch.list
  $ sudo apt-get update
  $ sudo apt-get -y install elasticsearch
  $ hfold --end

.. note:: To learn more about the ``hfold`` command see `Harrow Utilities -> hfold`_.

.. important::
  These instructions rely on a third-party distribution of ElasticSearch, the
  quality of which can hardly be tested. We recommend using the official
  Elasticsearch `Docker image`_.

.. _Docker image: ../../databases-and-services/elasticsearch/docker.html
.. _Harrow Utilities -> hfold: ../../harrow-utilities/hfold.html
