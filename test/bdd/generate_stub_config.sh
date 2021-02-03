#!/usr/bin/env bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

rm -rf ./fixtures/wellknown/jws/

mkdir -p ./fixtures/wellknown/jws/
mkdir -p ./fixtures/wellknown/jws/did-trustbloc/
 > ./fixtures/wellknown/jws/did-trustbloc/testnet.trustbloc.local.json
 > ./fixtures/wellknown/jws/did-trustbloc/stakeholder.one:8088.json
 > ./fixtures/wellknown/jws/did-trustbloc/stakeholder.two:8089.json

mkdir -p ./fixtures/wellknown/jws/genesis-configs
 > ./fixtures/wellknown/jws/genesis-configs/testnet.trustbloc.local.json

# directories that cli will write to
mkdir -p ./fixtures/wellknown/jws/stakeholder.one:8088/
mkdir -p ./fixtures/wellknown/jws/stakeholder.two:8089/
 > ./fixtures/wellknown/jws/stakeholder.one:8088/did-configuration.json
 > ./fixtures/wellknown/jws/stakeholder.two:8089/did-configuration.json

# source directories for docker container bind mounts
cp -r ./fixtures/wellknown/jws/stakeholder.one:8088 ./fixtures/wellknown/jws/stakeholder.one
cp -r ./fixtures/wellknown/jws/stakeholder.two:8089 ./fixtures/wellknown/jws/stakeholder.two
