# Harrow.io API

## Autoenv

This project makes use of [Autoenv] to assist with complex env variable
configuration. Please see the `env.sample` file, and consider copying it to
`.env` and using the Autoenv tool.

## Testing & Development

The project includes a `docker-compose.yml` file which can start the
dependencies enough to run the test suite.

    $ go get -u github.com/jteeuwen/go-bindata/...
    $ (cd src/github.com/harrowio/harrow && glide install)

## Installation

    $ docker-compose up [-d]
    $ make test

There's a way to run the tests individually

    $ time make t p=stores t=.                   # quiet
    $ time make tt p=stores t=Test_SessionStore  # verbose

The `t=` and `p=` are "tests to run" and "module" respectively, `p=` should be
set to the name of the package sans fully qualified url, i.e to run the tests
for `github.com/harrowio/harrow/stores` one would set `p=stores`. The `t=`
accepts the same regexp pattern as `-run`, see `go help testflag`.

## Running

Quite a few services (see the end of this document) must be running for Harrow
to function.

    $ # Install Caddy from https://caddyserver.com/download and make sure it's
        on the $PATH
    $ gem install foreman
    $ docker-compose up [-d]
    $ foreman start

Be mindful not to take a Procfile runner which autodetects a `.env` file, the
format is not compatible with `autoenv`, and will lead to problems such as Git
not being on the PATH because .env is assumed to be a file of key/value pairs
and will not be evaluated as it would be with autoenv.

For that reason `Foreman` rather than the Go port `forego` is recommended.

## TODO

- `activity_sink.go` `EmitActivity` has no logging possibility
- `RedisConnOpts` constants for keys/secrets source
- `authz.NewService` needs a logger

## Known Issues

Certain tests fail due to clock drift in Docker, for example:

    --- FAIL: Test_SessionStore_FindAllByUserUuid_sets_loaded_at_to_the_current_time (0.34s)
        session_store_test.go:176: loadedAtDelta = 4h36m12.419778053s; want < 5s

For this case, please restart the docker containers, or use another technique
for forcing the service to correct it's timestamp.

## Components

### Activity Worker

Monitors the `activities` table in the PostgreSQL database (via
zob/rabbitmq) and emits events to which other components respond. For
example when requesting a new password, an activity with the type
`user.requested-password-reset` will be emitted, this activity is used
to schedule an outbound email for the user containing their password
reset link.

### API

The HTTP/JSON API. This component interacts with the two data stores
(PostgreSQL and Redis) and also broadcasts activities to RabbitMQ (for example
when logging in, an activity is emitted). As a rule this component does not
emit events to which other components respond directly, instead relying on the
database triggers and the message bus to notify components which might be
interested in the results of changes.

### Build Status Worker

Reports build statuses to GitHub when conditions permit. Given a user on a
project who is permitted to post to a GitHub hosted repository when the build
status of the repository changes, we will optimistically post it to GitHub.

### Key Maker

Generates SSH keys by watching the activity bus to spot stub-keys which have
been named and saved, but have not had their secret bytes, or keypairs
generated yet. Upon successful generation is updates the primary store (Redis
in the case of keys) and also updates PostgreSQL notifying interested parties
that they key material is now present.

Key maker has three modes of operation `keymaker` ` keymaker rekey` `keymaker
rekey-missing`. The former listens and works, the `rekey` option destructively
rekeys the entire world, and the `rekey-missing` option catches up on missed
keys by generating key material for database stubs which are not yet correctly
generated.

### Mail Dispatcher

Mail dispatcher is responsible for sending the mails via the SMTP gateway. It
reads from a RabbitMQ queue upon which messages from the "Postal Worker" are
queueued.

### Metadata Preflight

This periodically scans all repositories which have Git triggers attached (i.e
"run task on new commits") and updates the PostgreSQL with the new metadata.
The presense of new metadata causes other events to be publishied which then
activate the scheduler, etc. Only repositories with associated Git triggers are
scanned. A one-time scan of all relevant repositories can be triggerd with the
`-once` flag, else the tool polls periodically (default 1 minute).

### Operation Runner

Runs operations queued into the database in a LXC container on a remote host.

The operation runner expects to be able to connect to an LXC/D host system
which has an image loaded with the alias `harrow-baseimage` (which  can be
retrieved from https://harrow-baseimage.s3.amazonaws.com/). The operator
requires SSH access to the LXC host. In development mode this is done by trying
to load `../config-management/.vagrant/machines/dev/virtualbox/private_key `.
Please note that this key will only exist if you've provisioned the development
virtual machine by following instructions in the `../config-management`
directory.

The operation runner spawns a subsystem (namely `controller-lxd`) which
actually manages the run. This runner is interchangeable.

### Postal Worker

Reacts to activities and other messages (namely the status changes of
operations) to determine which email message to send to which audience. This
coupled with the mail dispatcher completes the email sending subsystem.

### Projector

The projector builds optimized representations of certain objects (namely
"project cards") which would otherwise require large numbers of reads across
multiple tables and views on the database. The projector uses the activity
stream as published via RabbitMQ to invalidate and rebuild the cached objects.

### Scheduler

Scheduler looks for new "run once" schedules (e.g "now") to write the stub
operations to the database, it also looks for the recurring scheduled
(cronspecs, etc) to look for the correct time to publish a given operaion to
the databaes. The presence of the new stub operation in the database is enough
to trigger the operation-runner to

### WS (Web Socket)

The websocket server for the front-end events. This forwards relevant changes
on specific database table to subscribed and authenticated clients. The change
notification is used only to inform the front end SPA that the object should be
recalled using the normal HTTP request/response cycle.

### ZOB

The ZOB is a small layer to monitor PostgreSQL's non-persistent
[`NOTIFY`][https://www.postgresql.org/docs/current/static/sql-notify.html]
messages onto a more tolerant bus, in thie case RabbitMQ.

### Controller LXD / VMEx LXD

These two components start and stop LXC containers at the discretion of the
operation-runner on the LXC/D host in question. They need not be used directly.

## Running Harrow Without Limits / Billing

If you want to self-host Harrow you'll need to export the following:

    export HAR_LIMIT_STORE_FAIL_MODE=assume_allowed

This ensures that when the GRPC service for the limits subsystem fails, it will
fail "open", and no limits will be applied.

For the normal server this is sufficient, a failure in the limits package will
cancel the request context, which may or may not cause problems. For the tests
and other long-lived servers where the cancellation of a context may be fatal
(it's not possible to reuse a cancelled context) there is another option
available

    export HAR_LIMITS_ENABLED=false

The value given is parsed by Go's
[`strconv.ParseBool`](https://golang.org/pkg/strconv/#ParseBool).

*Note:* This double environment variable is less than ideal. A further look at
why the request context is cancelled in case of a bad `grpc.Dial` call would
probably allow us to get rid of the `HAR_LIMITS_ENABLED` env variable all
together.
