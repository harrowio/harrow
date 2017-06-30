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

## "Secret" Files

A handful of files in this repo are stored with the Ansible Vault function. To
keep these files out of Ansible's field of vision for the public released code
they are "stowed" using GNU stow.

Although it's rare that you would want to access these files, the system is simple:


    $ cd stow
    $ stow intern

Observe that this will create new files (symlinks) in `./inventories/` and
`./host_vars/`. If you are not in the list of authorized users Ansible will
fail because you are not able to decrypt the vault password file.

You can undo this process by manually removing the symlinks or running `stow -D
intern` from the `stow` directory.

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
