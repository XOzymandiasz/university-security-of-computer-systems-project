#!/bin/bash
set -e

mkdir -p pki/csr pki/issued pki/private pki/trust

echo "Generating TTP CA key"
openssl genrsa -out pki/private/ttp_ca.key 4096

echo "Generating TTP certificate"

openssl req -x509 -new -key pki/private/ttp_ca.key \
        -out pki/trust/ttp_ca.crt \
        -days 3650 \
        -subj "/CN=TTP"

echo "Generated"