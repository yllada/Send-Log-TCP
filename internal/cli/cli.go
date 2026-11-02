package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/syslog"
)

var Version = "dev"

const usage = `sendlog — send syslog messages to a remote server

Usage:
  sendlog                            GUI on Linux, terminal UI elsewhere
  sendlog tui                        Terminal UI (bubbletea)
  sendlog send [flags]               Send syslog messages
  sendlog connect [flags]            Test a connection
  sendlog stress [flags]             Continuous send at a fixed rate
  sendlog version                    Print the version
  sendlog help                       Show this help

send flags:
  -a address         Host or IP (required)
  -p port            Port, e.g. 514 or 6514 (required)
  -t protocol        tcp or udp (default tcp)
  --framing          octet-counting (default) or non-transparent (tcp only)
  --tls              Enable TLS (tcp only)
  -verify=false      Skip TLS certificate verification (default: verify)
  --ca path          CA certificate (PEM)
  --cert path        Client certificate for mTLS (PEM)
  --key path         Client key for mTLS (PEM)
  -m message         Message to send (repeatable)
  -f file            Log file to import (.json/.csv/.txt)
  --facility n       Facility 0-23 (default 16)
  --severity n       Severity 0-7 (default 6)
  --hostname name    Hostname field (default: this machine)
  --appname name     App-name field (default: sendlog)
  --format f         rfc5424 (default), rfc3164 or raw-pri
  --json             Print a single JSON result line to stdout

connect flags: same connection flags as send.

stress flags: connection + message flags, plus
  --rate n           Messages per second (default 10)
  --duration n       Stop after N seconds (0 = until Ctrl+C)
  --max n            Stop after N messages (0 = unlimited)

Progress and stats go to stderr; with --json one machine-readable line
goes to stdout. Exit code 0 on success, 1 on error.
`

type connFlags struct {
	address   string
	port      string
	protocol  string
	framing   string
	useTLS    bool
	tlsVerify bool
	ca        string
	cert      string
	key       string
	json      bool
}

func (c *connFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&c.address, "a", "", "host or IP")
	fs.StringVar(&c.address, "address", "", "host or IP")
	fs.StringVar(&c.port, "p", "", "port")
	fs.StringVar(&c.port, "port", "", "port")
	fs.StringVar(&c.protocol, "t", "tcp", "protocol: tcp or udp")
	fs.StringVar(&c.protocol, "protocol", "tcp", "protocol: tcp or udp")
	fs.StringVar(&c.framing, "framing", "", "tcp framing: octet-counting or non-transparent")
	fs.BoolVar(&c.useTLS, "tls", false, "enable TLS (tcp only)")
	fs.BoolVar(&c.tlsVerify, "verify", true, "verify TLS certificates")
	fs.BoolVar(&c.json, "json", false, "print a single JSON result line to stdout")
	fs.StringVar(&c.ca, "ca", "", "CA certificate path (PEM)")
	fs.StringVar(&c.cert, "cert", "", "client certificate path (PEM)")
	fs.StringVar(&c.key, "key", "", "client key path (PEM)")
}

type msgFlags struct {
	messages multiFlag
	file     string
	facility int
	severity int
	hostname string
	appname  string
	format   string
}

func (m *msgFlags) register(fs *flag.FlagSet) {
	fs.Var(&m.messages, "m", "message to send (repeatable)")
	fs.Var(&m.messages, "message", "message to send (repeatable)")
	fs.StringVar(&m.file, "f", "", "log file to import (.json/.csv/.txt)")
	fs.StringVar(&m.file, "file", "", "log file to import (.json/.csv/.txt)")
	fs.IntVar(&m.facility, "facility", 16, "facility 0-23")
	fs.IntVar(&m.severity, "severity", 6, "severity 0-7")
	fs.StringVar(&m.hostname, "hostname", "", "hostname field (default: this machine)")
	fs.StringVar(&m.appname, "appname", "sendlog", "app-name field")
	fs.StringVar(&m.format, "format", "rfc5424", "message format: rfc5424, rfc3164 or raw-pri")
}

func (m *msgFlags) messageFormat() syslog.MessageFormat {
	switch m.format {
	case "rfc5424":
		return syslog.MessageFormatRFC5424
	case "rfc3164":
		return syslog.MessageFormatRFC3164
	case "raw-pri":
		return syslog.MessageFormatRawPRI
	default:
		return syslog.MessageFormat(m.format)
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func connConfig(c *connFlags, m *msgFlags) syslog.Config {
	return syslog.Config{
		Address:        c.address,
		Port:           c.port,
		Protocol:       strings.ToLower(c.protocol),
		Method:         framing.Method(c.framing),
		Facility:       uint8(m.facility),
		Severity:       uint8(m.severity),
		Hostname:       m.hostname,
		Appname:        m.appname,
		MessageFormat:  m.messageFormat(),
		UseTLS:         c.useTLS,
		TLSVerify:      c.tlsVerify,
		CACertPath:     c.ca,
		ClientCertPath: c.cert,
		ClientKeyPath:  c.key,
	}
}

func emitJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot encode JSON result: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func Usage() string { return usage }
