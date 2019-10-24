#!/bin/bash

set -exuo pipefail


rm -f galley-*

certstrap --depot-path "." init --passphrase '' --key-bits 2048 --cn galley-ca
certstrap --depot-path "." request-cert --passphrase '' --ip '127.0.0.1' --cn galley-webhook
certstrap --depot-path "." sign --passphrase '' --CA galley-ca galley-webhook

rm -f *.crl *.csr
