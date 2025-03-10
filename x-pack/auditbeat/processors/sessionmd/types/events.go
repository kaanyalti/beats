// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package types

//go:generate stringer -linecomment=true -type=Type,HookPoint,Field -output=gen_types_string.go

import "github.com/elastic/beats/v7/auditbeat/helper/tty"

type Type uint64

const (
	ProcessFork Type = iota
	ProcessExec
	ProcessExit
	ProcessSetsid
)

type (
	Field uint32
)

const (
	CWD Field = iota + 1
	Argv
	Env
	Filename
)

type PIDInfo struct {
	StartTimeNS uint64
	Tid         uint32
	Tgid        uint32
	Vpid        uint32
	Ppid        uint32
	Pgid        uint32
	Sid         uint32
}

type CredInfo struct {
	Ruid         uint32
	Rgid         uint32
	Euid         uint32
	Egid         uint32
	Suid         uint32
	Sgid         uint32
	CapPermitted uint64
	CapEffective uint64
}

type ProcessForkEvent struct {
	ParentPIDs PIDInfo
	ChildPIDs  PIDInfo
	Creds      CredInfo
}

type ProcessExecEvent struct {
	PIDs             PIDInfo
	Creds            CredInfo
	CTTY             tty.TTYDev
	ProcfsLookupFail bool

	// varlen fields
	CWD      string
	Argv     []string
	Env      map[string]string
	Filename string
}

type ProcessExitEvent struct {
	PIDs     PIDInfo
	ExitCode int32
}

type ProcessSetsidEvent struct {
	PIDs PIDInfo
}
