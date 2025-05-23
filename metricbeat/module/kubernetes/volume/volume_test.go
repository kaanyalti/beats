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

//go:build !integration

package volume

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/elastic-agent-libs/logp/logptest"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

const testFile = "../_meta/test/stats_summary.json"

func TestEventMapping(t *testing.T) {
	logger := logptest.NewTestingLogger(t, "kubernetes.volume")

	f, err := os.Open(testFile)
	assert.NoError(t, err, "cannot open test file "+testFile)

	body, err := ioutil.ReadAll(f)
	assert.NoError(t, err, "cannot read test file "+testFile)

	events, err := eventMapping(body, logger)
	assert.NoError(t, err, "error mapping "+testFile)

	assert.Len(t, events, 2, "got wrong number of events")

	testCases := []map[string]interface{}{
		// Test for ephemeral volume
		{
			"name":               "default-token-sg8x5",
			"fs.available.bytes": 1939689472,
			"fs.capacity.bytes":  1939701760,
			"fs.used.bytes":      12288,
			"fs.used.pct":        float64(12288) / float64(1939701760),
			"fs.inodes.used":     9,
			"fs.inodes.free":     473551,
			"fs.inodes.count":    473560,
			"fs.inodes.pct":      float64(9) / float64(473560),
		},
		// Test for the persistent volume claim
		{
			mb.ModuleDataKey + ".persistentvolumeclaim.name": "pvc-demo",
			"name": "pvc-demo-vol",
		},
	}

	for i := range testCases {
		for k, v := range testCases[i] {
			testValue(t, events[i], k, v)
		}
	}
}

func testValue(t *testing.T, event mapstr.M, field string, value interface{}) {
	data, err := event.GetValue(field)
	assert.NoError(t, err, "Could not read field "+field)
	assert.EqualValues(t, data, value, "Wrong value for field "+field)
}
