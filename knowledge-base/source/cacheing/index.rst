Caching
========

Filesystem caching allows Operations_ to reuse pieces of the filesystem between
operations.

.. note::
  This feature is still in testing, and will be made generally available soon, or
  upon request. Simply `contact us`_.

The caching can be specified per-directory, and optionally given a cache key.
If the cache key is specified it should be something unique, otherwise the full
path will be used as the cache key.

The usage is very simple:

.. code-block:: bash

  $ hcache vendor/bundle

This will attempt to restore the directory at the path given from the
filesystem cache, this cache namespace is Project_ specific, that means other
Projects in your Organization_ can use the same key, without risking
overwriting, or fetching a bad cache.

**If the cache does not exist on the cache server** then a request will be made, to
cache anything in that path for future runs, at the end of the operation.

**If the cache exists on the cache server, but the local directory is not
empty** the command will fail, use ``hcache --force ./path``, in this case.

Advanced Usage
--------------

In certain situations it may be preferable have more control over the cache,
that is to say to be able to force the cache namespace, for sharing caches
between projects under a single organization, or by associating a more
restrictive cache key, or specifying a cache TTL.

.. code-block:: bash

  $ hcache vendor/bundle                              \
    [--cache-key=vendor/bundle]                       \
    [--ttl=30d]                                       \
    [--namespace=de305d54-75b4-431b-adb2-eb6b9e546014]
    [--force]

These are the default options.

The namespace option, where ``--namespace=de305...14`` is the current project
UUID. This value must be the current project UUID, or the organization UUID
which contains this project.  Other values will cause an exception, stop your
job and print a warning with an explanation.

.. warning::
  Overwriting the project namespace may violate the security of projects,
  allowing collaborators invited to one project access to see filesystem data
  from another. This is only a risk if your members are un-trusted, and the
  breadth of the leak is still constrained to the organization.

The ``--cache-key=...`` option can be used, for example to make restoring a
cached directory dependent upon a checksum or other measure:

.. code-block:: bash

  $ hcache vendor/bundle --cache-key=vendor-bundle-lock-$(md5sum Bundle.lock)
  $ hcache vendor/bundle --cache-key=vendor-bundle-rubyver-$(ruby -v)

The ``--ttl=`` argument accepts anything valid for Go's ```ParseDuration```_,
if a cache is not **used**, *or updated* for the ``--ttl`` value given, it will be
marked for cleanup, and eventually removed. Whilst occasionally cached values
may outlive their TTL due to the nature of the implementation we advise against
relying upon that.

A more mobust mechansm than ``--ttl=`` is to make the window the cache should
available part of the cache key:

.. code-block:: bash

  $ hcache vendor/bundle --cache-key=vendor-bundle-$(date +"%y-%m")

.. _ParseDuration: https://golang.org/pkg/time/#ParseDuration
.. _Operations: ../glossary#operations
.. _Project: ../glossary#project
.. _Organization: ../glossary#organization
.. _contact us: https://www.harrow.io/contact
