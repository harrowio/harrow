Using Ruby in Harrow via the official Docker container
=========================================================

Using the `official Ruby Docker containers`_ is one of the fastest and most
reliable ways to use Ruby on Harrow.

The containers are pulled directly each time, which can be time consuming (a
few seconds), but will eventually be cached in a Harrow-local Docker repository
mirror.

.. code-block:: bash

  #!/bin/bash -e

  echo 'puts "Hello Ruby"' | tee > example-script.rb
  sudo docker run -v "$PWD":/usr/src/example -w /usr/src/example ruby:latest ruby example-script.rb

To give a more concrete example, imagine you have a Ruby project checked out to
``repositories/my-project`` you might do something like the following:

.. code-block:: bash

  #!/bin/bash -e

  # In an ideal world, this script would be a part of your repository
  cat > repositories/my-project/lib/tasks/ci-deployment.rake <<-EOF
    namespace :harrow do
      desc "Perform the deployment on Harrow.io"
      task :deploy do
        sh "bundle install"
        sh "bundle exec cap #{ENV['STAGE']} deploy"
      end
    end
  EOF
  sudo docker run -v "$PWD":/my-project -w /my-project ruby:latest rake -f lib/tasks/ci-deployment.rake harrow:deploy


.. _official Ruby Docker containers: https://hub.docker.com/_/ruby/
