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

//go:build mage

package main

import (
	"fmt"
	"time"

	"github.com/magefile/mage/mg"
	"go.uber.org/multierr"

	auditbeat "github.com/elastic/beats/v7/auditbeat/scripts/mage"
	devtools "github.com/elastic/beats/v7/dev-tools/mage"
	"github.com/elastic/beats/v7/dev-tools/mage/target/build"

	//mage:import
	"github.com/elastic/beats/v7/dev-tools/mage/target/common"
	//mage:import
	"github.com/elastic/beats/v7/dev-tools/mage/target/unittest"
	//mage:import
	"github.com/elastic/beats/v7/dev-tools/mage/target/integtest"
	//mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/integtest/docker"
	//mage:import
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/test"
)

func init() {
	common.RegisterCheckDeps(Update)
	unittest.RegisterGoTestDeps(fieldsYML)
	integtest.RegisterGoTestDeps(fieldsYML)
	integtest.RegisterPythonTestDeps(Dashboards)

	devtools.BeatDescription = "Audit the activities of users and processes on your system."
}

// Build builds the Beat binary.
func Build() error {
	return devtools.Build(devtools.DefaultBuildArgs())
}

// GolangCrossBuild build the Beat binary inside of the golang-builder.
// Do not use directly, use crossBuild instead.
func GolangCrossBuild() error {
	return multierr.Combine(
		devtools.GolangCrossBuild(devtools.DefaultGolangCrossBuildArgs()),
		devtools.TestLinuxForCentosGLIBC(),
	)
}

// CrossBuild cross-builds the beat for all target platforms.
func CrossBuild() error {
	return devtools.CrossBuild()
}

// AssembleDarwinUniversal merges the darwin/amd64 and darwin/arm64 into a single
// universal binary using `lipo`. It assumes the darwin/amd64 and darwin/arm64
// were built and only performs the merge.
func AssembleDarwinUniversal() error {
	return build.AssembleDarwinUniversal()
}

// Package packages the Beat for distribution.
// Use SNAPSHOT=true to build snapshots.
// Use PLATFORMS to control the target platforms.
// Use VERSION_QUALIFIER to control the version qualifier.
func Package() {
	start := time.Now()
	defer func() { fmt.Println("package ran for", time.Since(start)) }()

	devtools.UseElasticBeatOSSPackaging()
	devtools.PackageKibanaDashboardsFromBuildDir()
	auditbeat.CustomizePackaging(auditbeat.OSSPackaging)

	mg.SerialDeps(Fields, Dashboards, Config, devtools.GenerateModuleIncludeListGo)
	mg.Deps(CrossBuild)
	mg.SerialDeps(devtools.Package, TestPackages)
}

// Package packages the Beat for IronBank distribution.
//
// Use SNAPSHOT=true to build snapshots.
func Ironbank() error {
	fmt.Println(">> Ironbank: this module is not subscribed to the IronBank releases.")
	return nil
}

// TestPackages tests the generated packages (i.e. file modes, owners, groups).
func TestPackages() error {
	return devtools.TestPackages()
}

// Update is an alias for running fields, dashboards, config, includes.
func Update() {
	mg.SerialDeps(Fields, Dashboards, Config,
		devtools.GenerateModuleIncludeListGo, Docs)
}

// Config generates both the short/reference configs and populates the modules.d
// directory.
func Config() error {
	return devtools.Config(devtools.AllConfigTypes, auditbeat.OSSConfigFileParams(), ".")
}

// Fields generates fields.yml and fields.go files for the Beat.
func Fields() {
	mg.Deps(libbeatAndAuditbeatCommonFieldsGo, moduleFieldsGo)
	mg.Deps(fieldsYML)
}

// libbeatAndAuditbeatCommonFieldsGo generates a fields.go containing both
// libbeat and auditbeat's common fields.
func libbeatAndAuditbeatCommonFieldsGo() error {
	if err := devtools.GenerateFieldsYAML(); err != nil {
		return err
	}
	return devtools.GenerateAllInOneFieldsGo()
}

// moduleFieldsGo generates a fields.go for each module.
func moduleFieldsGo() error {
	return devtools.GenerateModuleFieldsGo("module")
}

// fieldsYML generates the fields.yml file containing all fields.
func fieldsYML() error {
	return devtools.GenerateFieldsYAML("module")
}

// ExportDashboard exports a dashboard and writes it into the correct directory.
//
// Required environment variables:
// - MODULE: Name of the module
// - ID:     Dashboard id
func ExportDashboard() error {
	return devtools.ExportDashboard()
}

// Dashboards collects all the dashboards and generates index patterns.
func Dashboards() error {
	return devtools.KibanaDashboards("module")
}

// Docs collects the documentation.
func Docs() {
	mg.Deps(auditbeat.ModuleDocs, auditbeat.FieldDocs)
}
