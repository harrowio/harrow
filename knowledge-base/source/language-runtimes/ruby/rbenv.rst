Using Ruby in Harrow via ``rbenv``
==================================

Using the rbenv_ offers the widest range of Rubies quickly installed from
binary packages where possible.

.. code-block:: bash

  #!/bin/bash -e

  echo 'puts "Hello From Ruby #{RUBY_VERSION}"' | tee > example-script.rb
  rbenv local 2.2.2
  ruby --version
  ruby example-script.rb
  rbenv versions # lists all versions installed, but of course you can install others.

.. _rbenv: https://hub.docker.com/_/ruby/
