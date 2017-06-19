# Frontend

## Installation

    $ npm install -g bower
    $ npm install
    $ bower install

## Testing

    $ npm test

## Usage

Before the front end can be sensibly used, the backend must be bought up, or
the corresponding config in the `config/webpack.dev.js` must be modified to
proxy to another instance of the server:

    $ npm start
    $ open "http://www.local.harrow.io:8181/

Run the unit tests:

    $ npm test
