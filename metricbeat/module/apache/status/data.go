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

package status

import (
	"bufio"
	"errors"
	"regexp"
	"strings"

	s "github.com/elastic/beats/v7/libbeat/common/schema"
	c "github.com/elastic/beats/v7/libbeat/common/schema/mapstrstr"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var (
	scoreboardRegexp = regexp.MustCompile(`(Scoreboard):\s+((_|S|R|W|K|D|C|L|G|I|\.)+)`)

	// This should match: "CPUSystem: .01"
	matchNumber = regexp.MustCompile(`(^[0-9a-zA-Z ]+):\s+(\d*\.?\d+)`)

	schema = s.Schema{
		"total_accesses":    c.Int("Total Accesses"),
		"total_kbytes":      c.Int("Total kBytes"),
		"requests_per_sec":  c.Float("ReqPerSec", s.Optional),
		"bytes_per_sec":     c.Float("BytesPerSec", s.Optional),
		"bytes_per_request": c.Float("BytesPerReq", s.Optional),
		"workers": s.Object{
			"busy": c.Int("BusyWorkers"),
			"idle": c.Int("IdleWorkers"),
		},
		"uptime": s.Object{
			"server_uptime": c.Int("ServerUptimeSeconds"),
			"uptime":        c.Int("Uptime"),
		},
		"cpu": s.Object{
			"load":            c.Float("CPULoad", s.Optional),
			"user":            c.Float("CPUUser"),
			"system":          c.Float("CPUSystem"),
			"children_user":   c.Float("CPUChildrenUser"),
			"children_system": c.Float("CPUChildrenSystem"),
		},
		"connections": s.Object{
			"total": c.Int("ConnsTotal", s.Optional),
			"async": s.Object{
				"writing":    c.Int("ConnsAsyncWriting", s.Optional),
				"keep_alive": c.Int("ConnsAsyncKeepAlive", s.Optional),
				"closing":    c.Int("ConnsAsyncClosing", s.Optional),
			},
		},
		"load": s.Object{
			"1":  c.Float("Load1", s.Optional),
			"5":  c.Float("Load5", s.Optional),
			"15": c.Float("Load15", s.Optional),
		},
	}

	// Schema used till apache 2.4.12
	schemaOld = s.Schema{
		"total_accesses":    c.Int("Total Accesses"),
		"total_kbytes":      c.Int("Total kBytes"),
		"requests_per_sec":  c.Float("ReqPerSec", s.Optional),
		"bytes_per_sec":     c.Float("BytesPerSec", s.Optional),
		"bytes_per_request": c.Float("BytesPerReq", s.Optional),
		"workers": s.Object{
			"busy": c.Int("BusyWorkers"),
			"idle": c.Int("IdleWorkers"),
		},
		"uptime": s.Object{
			"uptime": c.Int("Uptime"),
		},
		"cpu": s.Object{
			"load": c.Float("CPULoad", s.Optional),
		},
		"connections": s.Object{
			"total": c.Int("ConnsTotal", s.Optional),
			"async": s.Object{
				"writing":    c.Int("ConnsAsyncWriting", s.Optional),
				"keep_alive": c.Int("ConnsAsyncKeepAlive", s.Optional),
				"closing":    c.Int("ConnsAsyncClosing", s.Optional),
			},
		},
	}
)

func applySchema(event mapstr.M, fullEvent map[string]interface{}) error {
	applicableSchema := schema
	if _, found := fullEvent["ServerUptimeSeconds"]; !found {
		applicableSchema = schemaOld
	}
	_, errs := applicableSchema.ApplyTo(event, fullEvent)
	return errors.Join(errs...)
}

// Map body to MapStr
func eventMapping(scanner *bufio.Scanner, hostname string, logger *logp.Logger) (mapstr.M, error) {
	var (
		totalS          int
		totalR          int
		totalW          int
		totalK          int
		totalD          int
		totalC          int
		totalL          int
		totalG          int
		totalI          int
		totalDot        int
		totalUnderscore int
		totalAll        int
	)

	fullEvent := map[string]interface{}{}

	// Iterate through all events to gather data
	for scanner.Scan() {
		if match := matchNumber.FindStringSubmatch(scanner.Text()); len(match) == 3 {
			// Total Accesses: 16147
			// Total kBytes: 12988
			// Uptime: 3229728
			// CPULoad: .000408393
			// CPUUser: 0
			// CPUSystem: .01
			// CPUChildrenUser: 0
			// CPUChildrenSystem: 0
			// ReqPerSec: .00499949
			// BytesPerSec: 4.1179
			// BytesPerReq: 823.665
			// BusyWorkers: 1
			// IdleWorkers: 8
			// ConnsTotal: 4940
			// ConnsAsyncWriting: 527
			// ConnsAsyncKeepAlive: 1321
			// ConnsAsyncClosing: 2785
			// ServerUptimeSeconds: 43
			//Load1: 0.01
			//Load5: 0.10
			//Load15: 0.06
			fullEvent[match[1]] = match[2]

		} else if match := scoreboardRegexp.FindStringSubmatch(scanner.Text()); len(match) == 4 {
			// Scoreboard Key:
			// "_" Waiting for Connection, "S" Starting up, "R" Reading Request,
			// "W" Sending Reply, "K" Keepalive (read), "D" DNS Lookup,
			// "C" Closing connection, "L" Logging, "G" Gracefully finishing,
			// "I" Idle cleanup of worker, "." Open slot with no current process
			// Scoreboard: _W____........___...............................................................................................................................................................................................................................................

			totalUnderscore = strings.Count(match[2], "_")
			totalS = strings.Count(match[2], "S")
			totalR = strings.Count(match[2], "R")
			totalW = strings.Count(match[2], "W")
			totalK = strings.Count(match[2], "K")
			totalD = strings.Count(match[2], "D")
			totalC = strings.Count(match[2], "C")
			totalL = strings.Count(match[2], "L")
			totalG = strings.Count(match[2], "G")
			totalI = strings.Count(match[2], "I")
			totalDot = strings.Count(match[2], ".")
			totalAll = totalUnderscore + totalS + totalR + totalW + totalK + totalD + totalC + totalL + totalG + totalI + totalDot
		} else {
			logger.Named("apache-status").Debugf("Unexpected line in apache server-status output: %s", scanner.Text())
		}
	}

	event := mapstr.M{
		"hostname": hostname,
		"scoreboard": mapstr.M{
			"starting_up":            totalS,
			"reading_request":        totalR,
			"sending_reply":          totalW,
			"keepalive":              totalK,
			"dns_lookup":             totalD,
			"closing_connection":     totalC,
			"logging":                totalL,
			"gracefully_finishing":   totalG,
			"idle_cleanup":           totalI,
			"open_slot":              totalDot,
			"waiting_for_connection": totalUnderscore,
			"total":                  totalAll,
		},
	}

	return event, applySchema(event, fullEvent)
}

/*
func parseMatchFloat(input interface{}, fieldName string) float64 {
	var parseString string

	if input != nil {
		if strings.HasPrefix(input.(string), ".") {
			parseString = strings.Replace(input.(string), ".", "0.", 1)
		} else {
			parseString = input.(string)
		}

		outputFloat, err := strconv.ParseFloat(parseString, 64)
		if err != nil {
			logp.Err("Cannot parse string '%s' to float for field '%s'. Error: %+v", input.(string), fieldName, err)
			return 0.0
		}
		return outputFloat
	} else {
		return 0.0
	}
}*/
