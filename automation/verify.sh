#!/usr/bin/env bash
set -euo pipefail
export ANSIBLE_NOCOLOR=1
expected=$'mode=healthy\nreplicas=1'
ansible-playbook -i localhost, /lab/ansible/site.yml
test "$(cat /tmp/aether-ansible/workload.conf)" = "$expected"
ansible-playbook -i localhost, /lab/ansible/site.yml | tee /tmp/ansible-second.log
grep -Eq 'changed=0 .*failed=0' /tmp/ansible-second.log
printf 'mode=degraded\n' >/tmp/aether-ansible/workload.conf
ansible-playbook -i localhost, /lab/ansible/site.yml
test "$(cat /tmp/aether-ansible/workload.conf)" = "$expected"
puppet apply --confdir /tmp/puppet-conf --vardir /tmp/puppet-var --rundir /tmp/puppet-run --detailed-exitcodes /lab/puppet/site.pp || test "$?" = 2
test "$(cat /tmp/aether-puppet/workload.conf)" = "$expected"
puppet apply --confdir /tmp/puppet-conf --vardir /tmp/puppet-var --rundir /tmp/puppet-run --detailed-exitcodes /lab/puppet/site.pp
printf 'mode=degraded\n' >/tmp/aether-puppet/workload.conf
puppet apply --confdir /tmp/puppet-conf --vardir /tmp/puppet-var --rundir /tmp/puppet-run --detailed-exitcodes /lab/puppet/site.pp || test "$?" = 2
test "$(cat /tmp/aether-puppet/workload.conf)" = "$expected"
echo "Ansible and Puppet converged, preserved idempotence and corrected drift"
