#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@did_method_rest
Feature: Using DID method REST API

  @e2e_universalresolver
  Scenario Outline: create trustbloc did and resolve through universalresolver
    Given Consortium config is generated with config file "./fixtures/wellknown/config.json"
    Then Discovery services are restarted
    Then TrustBloc DID is created through registrar "http://localhost:9080/1.0/register?driverId=driver-did-method-rest" with key type "<keyType>" with signature suite "<signatureSuite>"
    Then Bloc VDRI is initialized with resolver URL "http://localhost:8080/1.0/identifiers"
    Then Resolve created DID and validate key type "<keyType>", signature suite "<signatureSuite>"
    Examples:
      | keyType  |  signatureSuite             |
      | Ed25519  |  Ed25519VerificationKey2018 |

  @e2e_sidetree
  Scenario Outline: create trustbloc did and resolve through sidetree-mock
    Given Consortium config is generated with config file "./fixtures/wellknown/config.json"
    Then Discovery services are restarted
    Then TrustBloc DID is created through registrar "http://localhost:9080/1.0/register?driverId=driver-did-method-rest" with key type "<keyType>" with signature suite "<signatureSuite>"
    Then Bloc VDRI is initialized with resolver URL "https://localhost:48326/sidetree/0.0.1/identifiers"
    Then Resolve created DID and validate key type "<keyType>", signature suite "<signatureSuite>"
    Examples:
      | keyType  |  signatureSuite             |
      | Ed25519  |  JwsVerificationKey2020     |
      | P256     |  JwsVerificationKey2020     |
      | Ed25519  |  Ed25519VerificationKey2018 |

  @e2e_update_config_local
  Scenario Outline: create trustbloc did and resolve through local VDR after updating consortium config
    Given Consortium config is generated with config file "./fixtures/wellknown/stale_config.json"
    Then Discovery services are restarted
    Then TrustBloc DID is created through registrar "http://localhost:9080/1.0/register?driverId=driver-did-method-rest" with key type "<keyType>" with signature suite "<signatureSuite>"
    Then Bloc VDRI is initialized with genesis file "./fixtures/wellknown/jws/did-trustbloc/testnet.trustbloc.local.json"
    Then Consortium config is updated with config file "./fixtures/wellknown/update_config.json"
    Then Discovery services are restarted
    Then Resolve created DID and validate key type "<keyType>", signature suite "<signatureSuite>"
    Examples:
      | keyType  |  signatureSuite             |
      | Ed25519  |  JwsVerificationKey2020     |
      | P256     |  JwsVerificationKey2020     |
      | Ed25519  |  Ed25519VerificationKey2018 |

  @e2e_update_config_rest
  Scenario Outline: create trustbloc did and resolve through did-method-rest while using a genesis file
    Given Consortium config is generated with config file "./fixtures/wellknown/stale_config_rest.json"
    Then DID method service is restarted with genesis file "/etc/genesis-configs/testnet.trustbloc.local.json"
    Then TrustBloc DID is created through registrar "http://localhost:9080/1.0/register?driverId=driver-did-method-rest" with key type "<keyType>" with signature suite "<signatureSuite>"
    Then Bloc VDRI is initialized with resolver URL "http://localhost:8080/1.0/identifiers"
    Then Consortium config is updated with config file "./fixtures/wellknown/update_config_rest.json"
    Then Discovery services are restarted
    Then Resolve created DID and validate key type "<keyType>", signature suite "<signatureSuite>"
    Examples:
      | keyType  |  signatureSuite             |
      | Ed25519  |  JwsVerificationKey2020     |
      | P256     |  JwsVerificationKey2020     |
      | Ed25519  |  Ed25519VerificationKey2018 |

  @e2e_bad_update_config_local
  Scenario Outline: fail to resolve did through local VDR after a consortium config update isn't signed by stakeholders
    Given Consortium config is generated with config file "./fixtures/wellknown/stale_config.json"
    Then Discovery services are restarted
    Then TrustBloc DID is created through registrar "http://localhost:9080/1.0/register?driverId=driver-did-method-rest" with key type "<keyType>" with signature suite "<signatureSuite>"
    Then Bloc VDRI is initialized with genesis file "./fixtures/wellknown/jws/did-trustbloc/testnet.trustbloc.local.json"
    Then Consortium config is updated with config file "./fixtures/wellknown/bad_update_config.json"
    Then Discovery services are restarted
    Then DID resolution fails, containing error "config update signature does not verify"
    Examples:
      | keyType  |  signatureSuite             |
      | Ed25519  |  JwsVerificationKey2020     |
      | P256     |  JwsVerificationKey2020     |
      | Ed25519  |  Ed25519VerificationKey2018 |

  @e2e_bad_update_config_rest
  Scenario Outline: fail to resolve did through did-method-rest after a consortium config update isn't signed by stakeholders
    Given Consortium config is generated with config file "./fixtures/wellknown/stale_config_rest.json"
    Then DID method service is restarted with genesis file "/etc/genesis-configs/testnet.trustbloc.local.json"
    Then TrustBloc DID is created through registrar "http://localhost:9080/1.0/register?driverId=driver-did-method-rest" with key type "<keyType>" with signature suite "<signatureSuite>"
    Then Bloc VDRI is initialized with resolver URL "http://localhost:8080/1.0/identifiers"
    Then Consortium config is updated with config file "./fixtures/wellknown/bad_update_config_rest.json"
    Then Discovery services are restarted
    Then DID resolution fails, containing error "config update signature does not verify"
    Examples:
      | keyType  |  signatureSuite             |
      | Ed25519  |  JwsVerificationKey2020     |
      | P256     |  JwsVerificationKey2020     |
      | Ed25519  |  Ed25519VerificationKey2018 |
