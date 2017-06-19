Using PostgreSQL On Harrow With A Native Linux Package
======================================================

Installing PostgreSQL natively is faster than using `the Docker method`_ but
less flexible. The Docker method allows a choice of options about which version
to use, and more flexibility around ports, configuration, data directories,
etc.

.. code-block:: bash

  hfold "Installing PostgreSQL (Native)"
  sudo apt-get update
  sudo apt-get -y install postgresql postgresql-contrib
  sudo -u postgres createuser --superuser myuser
  sudo -u postgres psql -c "ALTER USER myuser WITH PASSWORD 'mypassword';"
  hfold --end

.. note:: To learn more about the ``hfold`` command see `Harrow Utilities -> hfold`_.

To verify that the database was correctly installed, you can run this Ruby example:

.. code-block:: bash

  $ cat > test.rb << EOF
  #!/usr/bin/env ruby

  require 'rubygems'
  require 'pg'

  # Output a table of current connections to the DB
  conn = PG.connect( dbname: 'postgres', user: 'postgres', host: '127.0.0.1', port: 5432 )
  conn.exec( "SELECT * FROM pg_stat_activity" ) do |result|
    puts "     PID | User             | Query"
    result.each do |row|
      puts " %7d | %-16s | %s " % row.values_at('pid', 'usename', 'query')
    end
  end
  EOF

  $ gem install 'pg' --no-ri --no-rdoc
  $ ruby test.rb

.. _the Docker method: ../../databases-and-services/postgresql/docker.html
.. _Harrow Utilities -> hfold: ../../harrow-utilities/hfold.html
