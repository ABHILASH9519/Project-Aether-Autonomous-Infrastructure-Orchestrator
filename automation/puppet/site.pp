file { '/tmp/aether-puppet':
  ensure => directory,
  mode   => '0700',
}
file { '/tmp/aether-puppet/workload.conf':
  ensure  => file,
  content => "mode=healthy\nreplicas=1\n",
  mode    => '0600',
  require => File['/tmp/aether-puppet'],
}
