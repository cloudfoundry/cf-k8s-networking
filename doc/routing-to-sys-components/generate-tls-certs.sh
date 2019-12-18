#!/bin/bash

set -exuo pipefail

if [ "$#" -ne 2 ]; then
  echo 'Usage ./generate-tls-certs [CERT_NAME_PREFIX] [CERT_DOMAIN_NAME]'
  exit 1
fi

cert_common_name="$2"
cert_name_prefix="$1"

certstrap --depot-path "." init --passphrase '' --key-bits 2048 --cn "${cert_name_prefix}-ca"
certstrap --depot-path "." request-cert --passphrase '' --cn "$cert_common_name" --key "${cert_name_prefix}.key" --csr "${cert_name_prefix}.csr"
certstrap --depot-path "." sign --passphrase '' --CA "${cert_name_prefix}-ca" "${cert_name_prefix}"

rm -f *.crl *.csr
