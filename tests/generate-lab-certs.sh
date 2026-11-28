#!/usr/bin/env bash
set -euo pipefail
: "${AETHER_TLS_DIR:?set absolute path to a fresh private directory}"
[[ "$AETHER_TLS_DIR" == /* ]]
umask 077
mkdir -p "$AETHER_TLS_DIR"
chmod 700 "$AETHER_TLS_DIR"
cd "$AETHER_TLS_DIR"
test ! -e ca.key
openssl req -x509 -newkey rsa:2048 -nodes -days 2 -subj '/CN=AETHER disposable lab CA' -addext 'basicConstraints=critical,CA:TRUE' -addext 'keyUsage=critical,keyCertSign,cRLSign' -keyout ca.key -out ca.crt
for name in fixture executor untrusted; do
  openssl req -newkey rsa:2048 -nodes -subj "/CN=$name" -keyout "$name.key" -out "$name.csr"
  if [[ "$name" == fixture ]]; then
    printf 'basicConstraints=critical,CA:FALSE\nkeyUsage=critical,digitalSignature,keyEncipherment\nextendedKeyUsage=serverAuth\nsubjectAltName=DNS:fixture\n' > "$name.ext"
  else
    printf 'basicConstraints=critical,CA:FALSE\nkeyUsage=critical,digitalSignature\nextendedKeyUsage=clientAuth\nsubjectAltName=URI:spiffe://aether.local/%s\n' "$name" > "$name.ext"
  fi
  openssl x509 -req -in "$name.csr" -CA ca.crt -CAkey ca.key -CAcreateserial -days 2 -extfile "$name.ext" -out "$name.crt"
done
# Directory remains 0700. Individual bind mounts must be readable by
# the distroless UID; never mount the directory or CA signing key.
chmod 644 ca.crt fixture.crt fixture.key executor.crt executor.key untrusted.crt untrusted.key
