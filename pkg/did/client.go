/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"bytes"
	"crypto/ed25519"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	docdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	log "github.com/sirupsen/logrus"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"

	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/discovery/staticdiscovery"
	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/endpoint"
	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/selection/staticselection"
)

const (
	sha2_256            = 18
	recoveryRevealValue = "recoveryOTP"
	updateRevealValue   = "updateOTP"
)

type discovery interface {
	GetEndpoints(domain string) ([]*endpoint.Endpoint, error)
}

type selection interface {
	SelectEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error)
}

// Client for did bloc
type Client struct {
	discovery discovery
	selection selection
	client    *http.Client
	tlsConfig *tls.Config
}

// New return did bloc client
func New(opts ...Option) *Client {
	c := &Client{client: &http.Client{}, selection: staticselection.NewService()}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	c.client.Transport = &http.Transport{TLSClientConfig: c.tlsConfig}
	c.discovery = staticdiscovery.NewService(staticdiscovery.WithTLSConfig(c.tlsConfig))

	return c
}

// CreateDID create did doc
func (c *Client) CreateDID(domain string, opts ...CreateDIDOption) (*docdid.Doc, error) {
	if domain == "" {
		return nil, errors.New("domain is empty")
	}

	createDIDOpts := &CreateDIDOpts{}
	// Apply options
	for _, opt := range opts {
		opt(createDIDOpts)
	}

	endpoints, err := c.getEndpoints(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}

	req, err := c.buildSideTreeRequest(createDIDOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to build sidetree request: %w", err)
	}

	resDoc, err := c.sendCreateRequest(req, endpoints[0].URL)
	if err != nil {
		return nil, fmt.Errorf("failed to send create sidetree request: %w", err)
	}

	return resDoc, nil
}

func (c *Client) getEndpoints(domain string) ([]*endpoint.Endpoint, error) {
	endpoints, err := c.discovery.GetEndpoints(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to discover endpoints: %w", err)
	}

	selectedEndpoints, err := c.selection.SelectEndpoints(endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to select endpoints: %w", err)
	}

	if len(selectedEndpoints) == 0 {
		return nil, errors.New("list of endpoints is empty")
	}

	return selectedEndpoints, nil
}

// buildSideTreeRequest request builder for sidetree public DID creation
func (c *Client) buildSideTreeRequest(createDIDOpts *CreateDIDOpts) ([]byte, error) {
	publicKeys := createDIDOpts.publicKeys

	doc := &Doc{
		PublicKey: publicKeys,
		Service:   createDIDOpts.services,
	}

	docBytes, err := doc.JSONBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get document bytes : %s", err)
	}

	recoveryKey, err := c.getRecoveryKey(publicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get recovery key : %s", err)
	}

	req, err := helper.NewCreateRequest(&helper.CreateRequestInfo{
		OpaqueDocument:          string(docBytes),
		RecoveryKey:             recoveryKey,
		NextRecoveryRevealValue: []byte(recoveryRevealValue),
		NextUpdateRevealValue:   []byte(updateRevealValue),
		MultihashCode:           sha2_256,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create sidetree request: %w", err)
	}

	return req, nil
}

func (c *Client) getRecoveryKey(publicKeys []PublicKey) (*jws.JWK, error) {
	for _, v := range publicKeys {
		if v.Recovery {
			if v.Encoding != PublicKeyEncodingJwk {
				return nil, fmt.Errorf("public key encoding not supported: %s", v.Encoding)
			}

			return pubkey.GetPublicKeyJWK(ed25519.PublicKey(v.Value))
		}
	}

	return nil, fmt.Errorf("recovery key not found")
}

func (c *Client) sendCreateRequest(req []byte, endpointURL string) (*docdid.Doc, error) {
	httpReq, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer closeResponseBody(resp.Body)

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response : %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got unexpected response from %s status '%d' body %s",
			endpointURL, resp.StatusCode, responseBytes)
	}

	didDoc, err := docdid.ParseDocument(responseBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public DID document: %s", err)
	}

	return didDoc, nil
}

func closeResponseBody(respBody io.Closer) {
	e := respBody.Close()
	if e != nil {
		log.Errorf("Failed to close response body: %v", e)
	}
}

// Option is a DID client instance option
type Option func(opts *Client)

// WithTLSConfig option is for definition of secured HTTP transport using a tls.Config instance
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(opts *Client) {
		opts.tlsConfig = tlsConfig
	}
}

// CreateDIDOpts create did opts
type CreateDIDOpts struct {
	publicKeys []PublicKey
	services   []docdid.Service
}

// CreateDIDOption is a create DID option
type CreateDIDOption func(opts *CreateDIDOpts)

// WithPublicKey add DID public key
func WithPublicKey(publicKey *PublicKey) CreateDIDOption {
	return func(opts *CreateDIDOpts) {
		opts.publicKeys = append(opts.publicKeys, *publicKey)
	}
}

// WithService add service
func WithService(service *docdid.Service) CreateDIDOption {
	return func(opts *CreateDIDOpts) {
		opts.services = append(opts.services, *service)
	}
}
