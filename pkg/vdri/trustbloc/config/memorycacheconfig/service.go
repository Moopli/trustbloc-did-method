/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package memorycacheconfig

import (
	"errors"
	"fmt"
	"time"

	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/models"

	"github.com/bluele/gcache"
)

type config interface {
	GetConsortium(string, string) (*models.ConsortiumFileData, error)
	GetStakeholder(string, string) (*models.StakeholderFileData, error)
}

// ConfigService fetches consortium and stakeholder configs using a wrapped config service, caching results in-memory
type ConfigService struct {
	config config
	cCache gcache.Cache
	sCache gcache.Cache
}

// NewService create new ConfigService
func NewService(config config) *ConfigService {
	configService := &ConfigService{
		config: config,
	}

	configService.cCache = makeExpiryReplacementCache(configService.cCache,
		configService.getNewCacheable(func(url, domain string) (cacheable, error) {
			return configService.config.GetConsortium(url, domain)
		}))

	configService.sCache = makeExpiryReplacementCache(configService.sCache,
		configService.getNewCacheable(func(url, domain string) (cacheable, error) {
			return configService.config.GetStakeholder(url, domain)
		}))

	return configService
}

func makeExpiryReplacementCache(cache gcache.Cache,
	fetcher func(url, domain string) (interface{}, *time.Duration, error)) gcache.Cache {
	return gcache.New(0).LoaderExpireFunc(func(key interface{}) (interface{}, *time.Duration, error) {
		keyStr, ok := key.(string)
		if !ok {
			return nil, nil, fmt.Errorf("key must be string")
		}

		return fetcher(keyStr, keyStr)
	}).Build()
}

type cacheable interface {
	CacheLifetime() (time.Duration, error)
}

func (cs *ConfigService) getNewCacheable(
	fetcher func(url, domain string) (cacheable, error)) func(url, domain string) (interface{}, *time.Duration, error) {
	return func(url, domain string) (interface{}, *time.Duration, error) {
		data, err := fetcher(url, domain)
		if err != nil {
			return nil, nil, fmt.Errorf("fetching cacheable object: %w", err)
		}

		expiryTime, err := data.CacheLifetime()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get object expiry time: %w", err)
		}

		return data, &expiryTime, nil
	}
}

func getEntryHelper(cache gcache.Cache, key interface{}, objectName string) (interface{}, error) {
	data, err := cache.Get(key)
	if err != nil {
		if errors.Is(err, gcache.KeyNotFoundError) {
			data, err = cache.Get(key)
			if err != nil {
				// if we got a KeyNotFoundError the first time, then it was expired
				// and it should have been added by the expired handler, and should be present now
				return nil, fmt.Errorf("refreshing expired %s %s", objectName, key)
			}
		} else {
			return nil, fmt.Errorf("getting %s from cache: %w", objectName, err)
		}
	}

	return data, nil
}

// GetConsortium fetches and parses the consortium file at the given domain, caching the value
func (cs *ConfigService) GetConsortium(url, domain string) (*models.ConsortiumFileData, error) {
	consortiumDataInterface, err := getEntryHelper(cs.cCache, url, "consortium")
	if err != nil {
		return nil, err
	}

	return consortiumDataInterface.(*models.ConsortiumFileData), nil
}

// GetStakeholder returns the stakeholder config file fetched by the wrapped config service, caching the value
func (cs *ConfigService) GetStakeholder(url, domain string) (*models.StakeholderFileData, error) {
	stakeholderDataInterface, err := getEntryHelper(cs.sCache, url, "stakeholder")
	if err != nil {
		return nil, err
	}

	return stakeholderDataInterface.(*models.StakeholderFileData), nil
}
