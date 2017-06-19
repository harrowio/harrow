# Harrow.io Open Source

<a href="https://cla-assistant.io/harrowio/harrow"><img src="https://cla-assistant.io/readme/badge/harrowio/harrow" alt="CLA assistant" /></a>

## PREVIEW

This is a brutally modified version of the upstream code which powers harrow.io
which is currently offered as a dev bundle for modifying Harrow. The intention
is that within a few weeks this message will go away, and harrow.io (hosted)
will run the same codebase as is offered here, but there is a small amount of
clean-up and canonicalization yet to do.

## What Is This?

Harrow is a task-runner for people who build and manage software. It's designed
to sit in the place of a traditional CI/CD build system whilst providing an
element of accessibility and beauty for non-technical team members and
stake-holders.

Harrow was borne out the popular Capistrano tool for Ruby (and Rails)
deployments and created by the same people.

Operating as a successful online SaaS since 2015, Harrow is now released in
it's entirety as a piece of (AGPL v3 licensed) free, open source software.

## Why Does This Exist?

Harrow sits in a peculiar place in DevOps. DevOps as a movement has reached a
plateau where "we have CI, and we do CD" has become the accepted state of the
art.

Harrow's creators believe that DevOps can go further, ideally we'd achieve the
same enlightenment that the Agile/XP movement brought to collaboration when
_building_ software and extend that to the whole life cycle of a piece of
probably business-critical software.

## Why Is The Code Open Source?

Harrow is/was VC funded, and having operated successfully so far, we want to
use the opportunity afforded to us to give something back to the FOSS community
to which we owe our existence.

## What Is Included?

The entire software is included. Some parts of the repository are protected
with GPG encryption, available only to those who are core maintainers of the
commercial, hosted version of Harrow.

There are some private components held in separate repositories, namely the
license key generation mechanisms for the enterprise version. The key
verification systems are part of this open source repository.

For a quick summary please see the following entries:

  * `frontend`: The Angular (1.x) application which drives our whole official
    HTML5 client. This application has extensive integration tests.
  * `style-guide`: Imported by the front end as a bower module, contains a
    separate style-guide with all graphic and styling resources.
  * `api`: The Go packages that comprise the application, including the
    `harrow` fat-executable which contains all the micro-services which fulfil
    all the roles responsible for the backend.
  * `notifiers`: The notification used by the API.
  * `knowledge-base`: The Sphinx based knowledge-base and deployment recipes.
  * `config-management`: Ansible scripts for building and provisioning the
    development, staging and test environments using VirtualBox. The same
    scripts are applied to production.

See the `README.md` of each subdirectory for more explanation of their contents.

## License

  * AGPL v3: https://www.gnu.org/licenses/agpl-3.0.html
  * See LICENSE.md

    Harrow, a continuous integration and collaboration software.

    Copyright (C) 2016 Harrow GmbH

    This program is free software: you can redistribute it and/or modify it
    under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or (at your
    option) any later version.

    This program is distributed in the hope that it will be useful, but WITHOUT
    ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
    FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public
    License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.

---
[Autoenv]: https://github.com/kennethreitz/autoenv
