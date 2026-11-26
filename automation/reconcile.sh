#!/bin/bash
set -euo pipefail
export ANSIBLE_NOCOLOR=1
mkdir -p "$HOME/.puppet/cache" /tmp/puppet-conf /tmp/puppet-var /tmp/puppet-run
ansible-playbook -i localhost, /lab/ansible/site.yml >/tmp/aether-ansible-reconcile.log
puppet apply --confdir /tmp/puppet-conf --vardir /tmp/puppet-var --rundir /tmp/puppet-run --detailed-exitcodes /lab/puppet/site.pp >/tmp/aether-puppet-reconcile.log || test "$?" = 2
