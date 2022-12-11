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
	"resenje.org/marshal"
)

// CompromisedOptions defines parameters related to service's core functionality.
type CompromisedOptions struct {
	// HTTP
	Listen                string            `json:"listen" yaml:"listen" envconfig:"LISTEN"`
	ListenInstrumentation string            `json:"listen-instrumentation" yaml:"listen-instrumentation" envconfig:"LISTEN_INSTRUMENTATION"`
	Headers               map[string]string `json:"headers" yaml:"headers" envconfig:"HEADERS"`
	RealIPHeaderName      string            `json:"real-ip-header-name" yaml:"real-ip-header-name" envconfig:"REAL_IP_HEADER_NAME"`
	// Passwords
	PasswordsDB string `json:"passwords-db" yaml:"passwords-db" envconfig:"PASSWORDS_DB"`
	// Logging
	LogDir string `json:"log-dir" yaml:"log-dir" envconfig:"LOG_DIR"`
	// Daemon
	DaemonLogFileName string       `json:"daemon-log-file" yaml:"daemon-log-file" envconfig:"DAEMON_LOG_FILE"`
	DaemonLogFileMode marshal.Mode `json:"daemon-log-file-mode" yaml:"daemon-log-file-mode" envconfig:"DAEMON_LOG_FILE_MODE"`
	PidFileName       string       `json:"pid-file" yaml:"pid-file" envconfig:"PID_FILE"`
}

// NewCompromisedOptions initializes CompromisedOptions with default values.
func NewCompromisedOptions() *CompromisedOptions {
	return &CompromisedOptions{
		Listen:                ":8080",
		ListenInstrumentation: "127.0.0.1:6060",
		Headers: map[string]string{
			"Server":          Name + "/" + compromised.Version(),
			"X-Frame-Options": "SAMEORIGIN",
		},
		RealIPHeaderName:  "X-Real-IP",
		PasswordsDB:       "",
		LogDir:            "",
		DaemonLogFileName: "daemon.log",
		DaemonLogFileMode: 0644,
		PidFileName:       filepath.Join(os.TempDir(), Name+".pid"),
	}
}

// VerifyAndPrepare implements application.Options interface.
func (o *CompromisedOptions) VerifyAndPrepare() (err error) {
	ln, err := net.Listen("tcp", o.Listen)
	if err != nil {
		return
	}
	ln.Close()
	ln, err = net.Listen("tcp", o.ListenInstrumentation)
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
