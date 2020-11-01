# Compromised

[![Go](https://github.com/janos/compromised/workflows/Go/badge.svg)](https://github.com/janos/compromised/actions)
[![PkgGoDev](https://pkg.go.dev/badge/resenje.org/compromised)](https://pkg.go.dev/resenje.org/compromised)
[![NewReleases](https://newreleases.io/badge.svg)](https://newreleases.io/github/janos/compromised)

**Validate if a password has already been compromised with on-premises service.**

This service is meant for people and organizations that want to protect their users from using already compromised passwords without exposing any information (password hash or even a part of it) to a third-party service, such is https://haveibeenpwned.com/. The same dataset is used as on haveibeenpwned, but only locally, as it provides the complete dataset do be downloaded.

This service is created for and used in production by [NewReleases](https://newreleases.io).

For any online service, NIST [SP 800-63B guidelines](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-63b.pdf) state that user-provided passwords should be checked against existing data breaches.

This service provides a CLI interface to run an HTTP API service to validate if a specific password has been compromised and how many times.

Its initial setup is not trivial as it requires a database to be generated from a publicly available data collection, while providing various options to reduce the database size.

## Installation

Compromised service binaries have no external dependencies and can just be copied and executed locally.

Binary downloads of the Compromised service can be found on the [Releases page](https://github.com/janos/compromised/releases/latest).

To install on Linux:

```sh
wget https://github.com/janos/compromised/releases/latest/download/compromised-linux-amd64 -O /usr/local/bin/compromised
chmod +x /usr/local/bin/compromised
```

You may need additional privileges to write to `/usr/local/bin`, but the file can be saved at any location that you want.

Supported operating systems and architectures:

- macOS 64bit `darwin-amd64`
- Linux 64bit `linux-amd64`
- Linux 32bit `linux-386`
- Linux ARM 64bit `linux-arm64`
- Linux ARM 32bit `linux-armv6`
- Windows 64bit `windows-amd64`
- Windows 32bit `windows-386`

This tool is implemented using the [Go programming language](https://golang.org) and can also be installed by issuing a `go get` command:

```sh
go get -u resenje.org/compromised/cmd/compromised
```

## Usage

This service does not distribute any passwords or password hashes. It relies on the validity of data provided by https://haveibeenpwned.com/Passwords and provides command to generate a searchable database from that data.

It provides an HTTP server with a JSON-encoded API endpoint to be used to validate if a password has been compromised and how many times.

In order to use the service it is required to generate the database and then start the service by loading the database.

### Getting help

Descriptions of available commands and flags can be printed with:

```sh
compromised -h
```

```
USAGE

  compromised [options...] [command]

  Executing the program without specifying a command will start a process in
  the foreground and log all messages to stderr.

COMMANDS

  daemon
    Start program in the background.

  stop
    Stop program that runs in the background.

  status
    Display status of a running process.

  config
    Print configuration that program will load on start. This command is
    dependent of -config-dir option value.

  debug-dump
    Send to a running process USR1 signal to log debug information in the log.

  index-passwords
    Generate passwords database from pwned passwords sha1 file.

  version
    Print version to Stdout.

OPTIONS

  -config-dir string
        Directory that contains configuration files.
  -h    Show program usage.
```

And flags of the `index-passwords` command:

```sh
compromised index-passwords -h
```

```
USAGE

  index-passwords [input filename] [output directory]

OPTIONS

  -h    Show program usage.
  -hash-counting string
        Store approximate hash counts. Possible values: exact, approx, none. (default "exact")
  -min-hash-count uint
        Skip hashes with counts lower than specified with this flag. (default 1)
  -shard-count int
        Split hashes into a several files. Possible values: 1, 2, 4, 8, 16, 32, 64, 128, 256. (default 32)
```

### Indexing password hashes

Download Pwned passwords SHA1 ordered by hash 7z file from https://haveibeenpwned.com/Passwords. This file is several gigabytes long (version 6 is 10.1GB) so make sure that you have enough disk space.

```sh
wget https://downloads.pwnedpasswords.com/passwords/pwned-passwords-sha1-ordered-by-hash-v6.7z
```

Extract a textual file from the downloaded 7z archive. This file is roughly twice in size of 7z archive that contains it, around 24G for version 6. Feel free to remove the 7z archive.

Generate the database with the following command:

```sh
compromised index-passwords \
    pwned-passwords-sha1-ordered-by-hash-v6.txt \
    compromised-passwords-db
```

This command will read the content of `pwned-passwords-sha1-ordered-by-hash-v6.txt` file (make sure that you enter the correct path to it) and store indexes in fast searchable database in `compromised-passwords-db` directory. Command `index-passwords` will create the directory itself and it will stop execution if it already exists. It is expected that the database size is around 12GB.

By default, all hashes are stored and indexed into 32 files called shards. It is possible to reduce the database size with two optional CLI flags `--hash-counting` and `--min-hash-count`.

For example:

```sh
compromised index-passwords \
    --hash-counting approx \
    --min-hash-count 10 \
    pwned-passwords-sha1-ordered-by-hash-v6.txt \
    compromised-passwords
```

Flag `--hash-counting` with `approx` value stores approximate hash counts by having exact values for very small values of to around 17 and with the larger values less precise (with variance of around 5%), but close enough to make an estimation on password popularity. With this option, the complete database is 9.7GB large.

Flag `--hash-counting` with `none` value does not store hash counts and API always returns 1 for count of compromised passwords. With this option, the complete database is 9.3GB large.

Flag `--min-hash-count` receives a numerical value which filters out all password hashes which have less number of compromisations than specified. This way it is possible to reduce the size of the database by excluding less frequently used passwords. For example by `--min-hash-count 2` only excluding passwords with count 1, the database size is reduced to 7.6GB, or with `--min-hash-count 5` to 1.9GB, or with `--min-hash-count 10` to 800MB.

You can combine these two options according to available capacity and the level of security and information that you want to provide.

### Configuration

Service configuration is stored in configuration file `compromised.yaml` in `/etc/compromised` directory by default. You can change the directory with `--config-dir` flag:

```sh
compromised --config-dir /data/config/compromised
```

All available options and their default values can be printed with:

```sh
compromised config
```

```
# compromised
---
listen: :8080
listen-internal: 127.0.0.1:6060
headers:
  Server: compromised/0.1.0-6ed439e-dirty
  X-Frame-Options: SAMEORIGIN
passwords-db: ""
log-dir: ""
log-level: DEBUG
syslog-facility: ""
syslog-tag: compromised
syslog-network: ""
syslog-address: ""
access-log-level: DEBUG
access-syslog-facility: ""
access-syslog-tag: compromised-access
daemon-log-file: daemon.log
daemon-log-file-mode: "644"
pid-file: /var/folders/l4/tn9ytbgs5xx76lshwgx5bj1w0000gn/T/compromised.pid

# config directories
---
- /etc/compromised
- /Users/janos/Library/Application Support/compromised
```

#### Environment variables

The service can be configured with environment variables as well. Variable names can be constructed based on the keys in configuration files.

For variables in `compromised.yaml`, capitalize all letters, replace `-` with `_` and prepend `COMPROMISED_` prefix. For example, to set `passwords-db`, the environment variable is `COMPROMISED_PASSWORDS_DB`:

```sh
COMPROMISED_PASSWORDS_DB=/path/to/passwrods-db compromised
```

### Starting the service

Executing the program without specifying a command will start a process in the foreground and log all messages to stderr:

```sh
compromised
```

Service requires `passwords-db` directory to be specified:

```sh
cat /etc/compromised/compromised.yaml
```

```yaml
passwords-db: /data/storage/compromised/passwrods
```

To write logs to files on local filesystem:

```sh
cat /etc/compromised/compromised.yaml
```

```yaml
passwords-db: /data/storage/compromised/passwrods
log-dir: /data/log/compromised
```

Paths in configuration files are given only as examples.

### Running in the background

The service can be run in the background and managed by itself with commands:

```sh
compromised daemon
```

```sh
compromised status
```

```sh
compromised stop
```

Or you can choose a process manager to manage it. For example this is a systemd service file:

```
[Unit]
Description=Compromised
After=network.target

[Service]
ExecStart=/usr/local/bin/compromised
ExecStop=/bin/kill $MAINPID
KillMode=none
Restart=on-failure
RestartPreventExitStatus=255
LimitNOFILE=65536
PrivateTmp=true
NoNewPrivileges=true

[Install]
WantedBy=default.target
```

### Using the API

In order to minimize the exposure of passwords that are checked, only SHA1 hash of a password is accepted by the API.

First calculate the hash (use printf, not echo as echo is appending new line):

```sh
printf 12345678 | sha1
```

```
7c222fb2927d828af22f592134e8932480637c0d
```

Then make an HTTP request like this one.

```sh
curl http://localhost:8080/v1/passwords/7c222fb2927d828af22f592134e8932480637c0d
```

```json
{"compromised":true,"count":2996082}
```

Make sure that the port is the same as you configured it for the `listen` option.

Of if you choose a very strong password:

```sh
printf "my not compromised password" | sha1sum
```

```
d391477a0849048fc28e62850a25518d72afd013
```

Then the HTTP response will look like this:

```sh
curl http://localhost:8080/v1/passwords/d391477a0849048fc28e62850a25518d72afd013
```

```json
{"compromised":false}
```

### Internal API

Beside the main API, there is another API endpoint, by default available on port `6060` only on `localhost` which exposes some of the internal information about the service:

- Prometheus metrics `http://localhost:6060/metrics`
- Most basic health check endpoint `http://localhost:6060/status`
- Most basic JSON health check endpoint `http://localhost:6060/api/status`
- Go pprof `http://localhost:6060/debug/pprof/`

Internal API can be disabled with an empty value for `listen-internal` configuration option in `/etc/compromised/compromised.yaml`:

```yaml
listen-internal: ""
```

## Using the Go library

As this service is written in the Go programming language, an HTTP client package is provided, but also a package that allows loading the database in your own application if you do not want to manage the `compromised` service.

### HTTP Client

```go
package main

import (
	"contex"
	"crypto/sha1"
	"fmt"

	httppasswords "resenje.org/compromised/pkg/passwords/http"
)

func main() {
	// url with host and port where compromised service is listening
	s, err := httppasswords.New("http://localhost:8080", nil)
	if err != nil {
		panic(err)
	}

	c, err := s.IsPasswordCompromised(contex.Background(), sha1.Sum([]byte("my password")))
	if err != nil {
		panic(err)
	}

	fmt.Println("this password has been compromised", c, "times")
}
```

### Embed DB

```go
package main

import (
	"contex"
	"crypto/sha1"
	"fmt"

	filepasswords "resenje.org/compromised/pkg/passwords/file"
)

func main() {
	s, err := filepasswords.New("/path/to/passwords-db")
	if err != nil {
		panic(err)
	}
	defer s.Close()

	c, err := s.IsPasswordCompromised(contex.Background(), sha1.Sum([]byte("my password")))
	if err != nil {
		panic(err)
	}

	fmt.Println("this password has been compromised", c, "times")
}
```

## Versioning

To see the current version of the binary, execute:

```sh
compromised version
```

Each version is tagged and the version is updated accordingly in `version.go` file.

## Contributing

Read the [contribution guidelines](CONTRIBUTING.md).

## License

This application is distributed under the BSD-style license found in the [LICENSE](LICENSE) file.
