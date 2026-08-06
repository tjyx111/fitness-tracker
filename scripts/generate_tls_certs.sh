#!/usr/bin/env bash
set -euo pipefail

TLS_DIR="${FITNESS_TLS_DIR:-/root/.config/fitness-tracker/tls}"
SERVER_IP="${FITNESS_SERVER_IP:-111.230.63.109}"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required" >&2
  exit 1
fi

mkdir -p "$TLS_DIR"
chmod 700 "$TLS_DIR"

for file in ca.key ca.crt server.key server.crt client.key client.crt client.p12 client-p12-password.txt; do
  if [ -e "$TLS_DIR/$file" ]; then
    echo "Refusing to overwrite existing TLS file: $TLS_DIR/$file" >&2
    exit 1
  fi
done

umask 077

openssl req -x509 -newkey rsa:3072 -sha256 -nodes \
  -days 3653 \
  -subj "/CN=Assistant Local CA/O=Assistant" \
  -addext "basicConstraints=critical,CA:TRUE,pathlen:0" \
  -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -addext "subjectKeyIdentifier=hash" \
  -keyout "$TLS_DIR/ca.key" \
  -out "$TLS_DIR/ca.crt"

openssl req -new -newkey rsa:3072 -sha256 -nodes \
  -subj "/CN=assistant-server/O=Assistant" \
  -addext "subjectAltName=IP:${SERVER_IP},IP:127.0.0.1,DNS:localhost" \
  -addext "keyUsage=critical,digitalSignature,keyEncipherment" \
  -addext "extendedKeyUsage=serverAuth" \
  -keyout "$TLS_DIR/server.key" \
  -out "$TLS_DIR/server.csr"

openssl x509 -req \
  -in "$TLS_DIR/server.csr" \
  -CA "$TLS_DIR/ca.crt" \
  -CAkey "$TLS_DIR/ca.key" \
  -CAcreateserial \
  -days 3650 \
  -sha256 \
  -copy_extensions copy \
  -out "$TLS_DIR/server.crt"

openssl req -new -newkey rsa:3072 -sha256 -nodes \
  -subj "/CN=assistant-client/O=Assistant" \
  -addext "keyUsage=critical,digitalSignature" \
  -addext "extendedKeyUsage=clientAuth" \
  -keyout "$TLS_DIR/client.key" \
  -out "$TLS_DIR/client.csr"

openssl x509 -req \
  -in "$TLS_DIR/client.csr" \
  -CA "$TLS_DIR/ca.crt" \
  -CAkey "$TLS_DIR/ca.key" \
  -days 3650 \
  -sha256 \
  -copy_extensions copy \
  -out "$TLS_DIR/client.crt"

openssl rand -base64 -out "$TLS_DIR/client-p12-password.txt" 24
openssl pkcs12 -export \
  -inkey "$TLS_DIR/client.key" \
  -in "$TLS_DIR/client.crt" \
  -certfile "$TLS_DIR/ca.crt" \
  -name "Assistant Client" \
  -passout "file:$TLS_DIR/client-p12-password.txt" \
  -out "$TLS_DIR/client.p12"

rm -f "$TLS_DIR/server.csr" "$TLS_DIR/client.csr" "$TLS_DIR/ca.srl"
chmod 600 "$TLS_DIR"/*.key
chmod 600 "$TLS_DIR/client.p12" "$TLS_DIR/client-p12-password.txt"
chmod 644 "$TLS_DIR"/*.crt

openssl verify -CAfile "$TLS_DIR/ca.crt" -purpose sslserver "$TLS_DIR/server.crt"
openssl verify -CAfile "$TLS_DIR/ca.crt" -purpose sslclient "$TLS_DIR/client.crt"

echo "TLS certificates created in: $TLS_DIR"
echo "Server and client certificate validity: 3650 days"
