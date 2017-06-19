Using PostgreSQL On Harrow With Docker
======================================

The `official PostgreSQL Docker image`_ can easily be fetched into the
Harrow environment with the following command. Beware the image will not
persist any data (for information about persisting data see caching_).

.. code-block:: bash

  $ hfold "Starting PostgreSQL (Docker)"
  $ sudo docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=supersecret postgres:latest
  $ hfold --end

.. note:: To learn more about the `hfold` command see hfold_.

Be aware that PostgreSQL is installed in a Docker container, and as such the
client libraries and other parts of the package and it's dependencies may not
be available in the operation host machine, for example if your application
wishes to connect to PostgreSQL you'll need to install client libraries, here's
a Ruby example:

.. code-block:: bash

  $ cat > test.rb << EOF
  #!/usr/bin/env ruby

  require 'rubygems'
  require 'pg'

  # Output a table of current connections to the DB
  conn = PG.connect( password: 'supersecret', dbname: 'postgres', user: 'postgres', host: '127.0.0.1', port: 5432 )
  conn.exec( "SELECT * FROM pg_stat_activity" ) do |result|
    puts "     PID | User             | Query"
    result.each do |row|
      puts " %7d | %-16s | %s " % row.values_at('pid', 'usename', 'query')
    end
  end
  EOF

  $ gem install 'pg' --no-ri --no-rdoc
  $ ruby test.rb

.. _hfold: /harrow-utilities/hfold
.. _official PostgreSQL Docker image: https://hub.docker.com/_/postgres/
.. _caching: /caching/index
