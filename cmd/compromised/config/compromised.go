// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"net"
	"os"
	"path/filepath"

	"resenje.org/compromised"
	"resenje.org/logging"
	"resenje.org/marshal"
)

// CompromisedOptions defines parameters related to service's core functionality.
type CompromisedOptions struct {
	// HTTP
	Listen         string            `json:"listen" yaml:"listen" envconfig:"LISTEN"`
	ListenInternal string            `json:"listen-internal" yaml:"listen-internal" envconfig:"LISTEN_INTERNAL"`
	Headers        map[string]string `json:"headers" yaml:"headers" envconfig:"HEADERS"`
	// Passwords
	PasswordsDB string `json:"passwords-db" yaml:"passwords-db" envconfig:"PASSWORDS_DB"`
	// Logging
	LogDir               string                 `json:"log-dir" yaml:"log-dir" envconfig:"LOG_DIR"`
	LogLevel             logging.Level          `json:"log-level" yaml:"log-level" envconfig:"LOG_LEVEL"`
	SyslogFacility       logging.SyslogFacility `json:"syslog-facility" yaml:"syslog-facility" envconfig:"SYSLOG_FACILITY"`
	SyslogTag            string                 `json:"syslog-tag" yaml:"syslog-tag" envconfig:"SYSLOG_TAG"`
	SyslogNetwork        string                 `json:"syslog-network" yaml:"syslog-network" envconfig:"SYSLOG_NETWORK"`
	SyslogAddress        string                 `json:"syslog-address" yaml:"syslog-address" envconfig:"SYSLOG_ADDRESS"`
	AccessLogLevel       logging.Level          `json:"access-log-level" yaml:"access-log-level" envconfig:"ACCESS_LOG_LEVEL"`
	AccessSyslogFacility logging.SyslogFacility `json:"access-syslog-facility" yaml:"access-syslog-facility" envconfig:"ACCESS_SYSLOG_FACILITY"`
	AccessSyslogTag      string                 `json:"access-syslog-tag" yaml:"access-syslog-tag" envconfig:"ACCESS_SYSLOG_TAG"`
	// Daemon
	DaemonLogFileName string       `json:"daemon-log-file" yaml:"daemon-log-file" envconfig:"DAEMON_LOG_FILE"`
	DaemonLogFileMode marshal.Mode `json:"daemon-log-file-mode" yaml:"daemon-log-file-mode" envconfig:"DAEMON_LOG_FILE_MODE"`
	PidFileName       string       `json:"pid-file" yaml:"pid-file" envconfig:"PID_FILE"`
}

// NewCompromisedOptions initializes CompromisedOptions with default values.
func NewCompromisedOptions() *CompromisedOptions {
	return &CompromisedOptions{
		Listen:         ":8080",
		ListenInternal: "127.0.0.1:6060",
		Headers: map[string]string{
			"Server":          Name + "/" + compromised.Version,
			"X-Frame-Options": "SAMEORIGIN",
		},
		PasswordsDB:          "",
		LogDir:               "",
		LogLevel:             logging.DEBUG,
		SyslogFacility:       "",
		SyslogTag:            Name,
		SyslogNetwork:        "",
		SyslogAddress:        "",
		AccessLogLevel:       logging.DEBUG,
		AccessSyslogFacility: "",
		AccessSyslogTag:      Name + "-access",
		DaemonLogFileName:    "daemon.log",
		DaemonLogFileMode:    0644,
		PidFileName:          filepath.Join(os.TempDir(), Name+".pid"),
	}
}

// VerifyAndPrepare implements application.Options interface.
func (o *CompromisedOptions) VerifyAndPrepare() (err error) {
	ln, err := net.Listen("tcp", o.Listen)
	if err != nil {
		return
	}
	ln.Close()
	ln, err = net.Listen("tcp", o.ListenInternal)
	if err != nil {
		return
	}
	ln.Close()

	for _, dir := range []string{
		filepath.Dir(o.PidFileName),
		o.LogDir,
	} {
		if dir != "" {
			if err := os.MkdirAll(dir, 0777); err != nil {
				return err
			}
		}
	}
	return
}
