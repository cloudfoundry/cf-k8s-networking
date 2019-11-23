#!/bin/bash

set -exuo pipefail

cert_common_name='*.sys.eirini-dev-1.routing.lol'

rm -f sys-*

certstrap --depot-path "." init --passphrase '' --key-bits 2048 --cn sys-ca
certstrap --depot-path "." request-cert --passphrase '' --cn "$cert_common_name" --key sys-wildcard.key --csr sys-wildcard.csr
certstrap --depot-path "." sign --passphrase '' --CA sys-ca sys-wildcard

rm -f *.crl *.csr
