// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package activedirectory

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/elastic/elastic-agent-libs/transport/tlscommon"
)

// defaultConfig returns a default configuration.
func defaultConfig() conf {
	return conf{
		SyncInterval:   24 * time.Hour,
		UpdateInterval: 15 * time.Minute,
	}
}

// conf contains parameters needed to configure the input.
type conf struct {
	BaseDN string `config:"ad_base_dn" validate:"required"`

	URL      string `config:"ad_url" validate:"required"`
	User     string `config:"ad_user" validate:"required"`
	Password string `config:"ad_password" validate:"required"`

	// Dataset specifies the datasets to collect from
	// the API. It can be ""/"all", "users", or
	// "devices".
	Dataset string `config:"dataset"`
	// UserQuery and DeviceQuery are the LDAP queries
	// to obtain users and devices if set.
	// These are not currently intended to be used
	// but are included for future behavior expansion.
	UserQuery   string `config:"user_query"`
	DeviceQuery string `config:"device_query"`

	UserAttrs []string `config:"user_attributes"`
	GrpAttrs  []string `config:"group_attributes"`

	PagingSize uint32 `config:"ad_paging_size"`

	// SyncInterval is the time between full
	// synchronisation operations.
	SyncInterval time.Duration `config:"sync_interval"`
	// UpdateInterval is the time between
	// incremental updated.
	UpdateInterval time.Duration `config:"update_interval"`

	// TLS provides ssl/tls setup settings
	TLS *tlscommon.Config `config:"ssl" yaml:"ssl,omitempty" json:"ssl,omitempty"`
}

var (
	errInvalidSyncInterval   = errors.New("zero or negative sync_interval")
	errInvalidUpdateInterval = errors.New("zero or negative update_interval")
	errSyncBeforeUpdate      = errors.New("sync_interval not longer than update_interval")
)

// Validate runs validation against the config.
func (c *conf) Validate() error {
	switch {
	case c.SyncInterval <= 0:
		return errInvalidSyncInterval
	case c.UpdateInterval <= 0:
		return errInvalidUpdateInterval
	case c.SyncInterval <= c.UpdateInterval:
		return errSyncBeforeUpdate
	}
	switch strings.ToLower(c.Dataset) {
	case "", "all", "users", "devices":
	default:
		return errors.New("dataset must be 'all', 'users', 'devices' or empty")
	}
	_, err := ldap.ParseDN(c.BaseDN)
	if err != nil {
		return err
	}
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}
	if c.TLS.IsEnabled() && u.Scheme == "ldaps" {
		_, err := tlscommon.LoadTLSConfig(c.TLS)
		if err != nil {
			return err
		}
		_, _, err = net.SplitHostPort(u.Host)
		var addrErr *net.AddrError
		switch {
		case err == nil:
		case errors.As(err, &addrErr):
			if addrErr.Err != "missing port in address" {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

func (c *conf) wantUsers() bool {
	switch strings.ToLower(c.Dataset) {
	case "", "all", "users":
		return true
	default:
		return false
	}
}

func (c *conf) wantDevices() bool {
	switch strings.ToLower(c.Dataset) {
	case "", "all", "devices":
		return true
	default:
		return false
	}
}
