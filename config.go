// Copyright 2025 Acuvity, Inc.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

var _ component.Config = (*Config)(nil)
var _ xconfmap.Validator = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Config is the configuration for the Acuvity exporter
type Config struct {
	APIToken     string `mapstructure:"api_token"`
	APICA        string `mapstructure:"api_ca"`
	APIURL       string `mapstructure:"api_url"`
	APINamespace string `mapstructure:"api_namespace"`
}

// Validate implements xconfmap.Validator.
func (c *Config) Validate() error {

	if c.APIToken == "" {
		return errors.New("API token must be non-empty")
	}

	return nil
}

// Unmarshal a confmap.Conf into the config struct.
func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {

	if componentParser == nil {
		return errors.New("empty config for acuvity exporter")
	}

	err := componentParser.Unmarshal(cfg)
	if err != nil {
		return err
	}

	return nil
}
