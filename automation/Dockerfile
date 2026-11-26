FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ansible-core puppet-agent ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /lab
COPY ansible /lab/ansible
COPY puppet /lab/puppet
COPY verify.sh /lab/verify.sh
ENV ANSIBLE_LOCAL_TEMP=/tmp/ansible-local ANSIBLE_REMOTE_TEMP=/tmp/ansible-remote
ENTRYPOINT ["/bin/bash","/lab/verify.sh"]
