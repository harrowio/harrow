# Ansible

## Requirements

Owing to weirdness with the Ubuntu 16.04, VirtualBox and Vagrant
incompatibilities, the following versions of software are mandatory.

The specific known-good baseimage version is named in the `Vagrantfile`.

  - [VirtualBox](https://www.virtualbox.org/wiki/Downloads) >= 5.1
  - [Vagrant](https://www.vagrantup.com/downloads.html) >= 1.8

## Installation

If you don't want to use a virutalenv, please feel free to skip those steps if
you know what you're doing:

    # easy_install pip
    # pip install virtualenv
    $ virtualenv venv
    $ (source venv/bin/activate && pip install "ansible>=2.3,<2.4")

## Vagrant Plugins

    $ vagrant plugin install vagrant-persistent-storage

## Galaxy

    $ ansible-galaxy install -p roles -r ./requirements.yml

## Running

    $ make development
    # <make a coffee>

## TODO LXD

 * sudo apt-get install python3-dateutil
 * 'net.ipv4.ip_forward=1' in /etc/sysctl.conf (# sysctl -p)
