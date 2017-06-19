``hfold`` Collapse Regions Of Log Output
========================================

The ``hfold`` command is a shell utility helper which emits a specific `ANSI escape code`_.

The command can be used in two ways, first the shorthand:

.. code-block:: bash

  $ hfold "Installing Gem Bundle" bundle install

and, for multi line commands, the long-hand:

.. code-block:: bash

  $ hfold "Installing Node.js Dependencies"
  $ npm install -g karma-cli bower
  $ npm install
  $ bower install
  $ hfold --end

Use outside of the shell
------------------------

If you are writing tools or utilities in a language other than Bash, the escape
sequence has the following structure:

.. code-block:: bash

    $ printf "\033]10;\"%s\"\a" "Title Here" # opens a fold block
    $ printf "\033]10\a"                     # ends the fold block

.. _ANSI escape code: https://en.wikipedia.org/wiki/ANSI_escape_code
