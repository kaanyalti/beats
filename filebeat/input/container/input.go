// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package container

import (
	"errors"
	"fmt"

	"github.com/elastic/beats/v7/filebeat/channel"
	"github.com/elastic/beats/v7/filebeat/input"
	"github.com/elastic/beats/v7/filebeat/input/log"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var (
	errDeprecated = errors.New("Container input is deprecated. Use Filestream input with its container parser instead. https://www.elastic.co/guide/en/beats/filebeat/current/filebeat-input-filestream.html#_container") //nolint:staticcheck //Keep old behavior
)

func init() {
	err := input.Register("container", NewInput)
	if err != nil {
		panic(err)
	}
}

// NewInput creates a new container input
func NewInput(
	cfg *conf.C,
	outletFactory channel.Connector,
	context input.Context,
	logger *logp.Logger,
) (input.Input, error) {
	// we still allow the deprecated log input running under integrations and
	// modules until they are all migrated to filestream
	if !log.AllowDeprecatedUse(cfg) {
		return nil, fmt.Errorf("Found container input configuration: %w\n%s", errDeprecated, conf.DebugString(cfg, true)) //nolint:staticcheck //Keep old behavior
	}

	// Wrap log input with custom docker settings
	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("reading container input config: %w", err)
	}

	err := cfg.Merge(mapstr.M{
		"docker-json.partial":   true,
		"docker-json.cri_flags": true,

		// Allow stream selection (stdout/stderr/all)
		"docker-json.stream": config.Stream,

		// Select file format (auto/cri/docker)
		"docker-json.format": config.Format,

		// Set symlinks to true as CRI-O paths could point to symlinks instead of the actual path.
		"symlinks": true,
	})
	if err != nil {
		return nil, fmt.Errorf("update input config: %w", err)
	}

	// Add stream to meta to ensure different state per stream
	if config.Stream != "all" {
		if context.Meta == nil {
			context.Meta = map[string]string{}
		}
		context.Meta["stream"] = config.Stream
	}

	return log.NewInput(cfg, outletFactory, context, logger)
}
