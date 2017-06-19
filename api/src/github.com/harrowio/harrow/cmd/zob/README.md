# ZOB (Zentraler Omnibusbahnhof)

Routes pg create and change message to broadcast.create and
broadcast.change amqp channels.

## Usage:

    $ HARROW_ENV=development zob

## Requirements:

It requires that all messages published are JSON.

It won't inspect the bodies, but will add a `Content-Type` when forwarding to
RabbitMQ where the `Content-Type` may be important.
