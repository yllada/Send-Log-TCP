// Command sendlog-cli is a Linux command-line client for SendLog Syslog.
//
// It reuses the exact core of the desktop application (message building,
// RFC 6587 framing, TCP/UDP/TLS dialing, stress engine, log import and
// profile/template storage), so profiles created in the GUI are available
// here and vice versa.
//
// Progress and stats go to stderr; with --json a single machine-readable
// line is written to stdout.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/logimport"
	"Sendlog-Syslog/internal/storage"
	"Sendlog-Syslog/internal/stress"
	"Sendlog-Syslog/internal/syslog"
)

const usageText = `sendlog-cli — SendLog Syslog command-line client (Linux)

Usage:
  sendlog-cli send [flags]                 Send syslog messages
  sendlog-cli connect [flags]              Test a connection
  sendlog-cli stress [flags]               Continuous/stress send (Ctrl+C to stop)
  sendlog-cli profile list                 List saved profiles
  sendlog-cli profile add [flags]          Save a connection profile
  sendlog-cli profile remove <id>          Delete a profile
  sendlog-cli template list                List saved templates
  sendlog-cli import -f <file> [--merge]   Import profiles/templates from JSON
  sendlog-cli export                       Print profiles/templates as JSON
  sendlog-cli version

Connection flags (send/connect):
  -a address        Host or IP (required)
  -p port           Port, e.g. 514 or 6514 (required)
  -t protocol       tcp or udp (default tcp)
  --framing         octet-counting (default) or non-transparent (tcp only)
  --tls             Enable TLS (tcp only, RFC 5425)
  -verify=false      Skip TLS certificate verification (default: verify)
  --ca path         Custom CA certificate (PEM)
  --cert path       Client certificate for mTLS (PEM)
  --key path        Client key for mTLS (PEM)

Message flags (send, stress):
  -m message        Message to send (repeatable); for stress, the base content
  -f file           Log file to import (.json/.csv/.txt, like the GUI batch import)
  --facility n      Facility 0-23 (default 16/local0)
  --severity n      Severity 0-7 (default 6/info)
  --hostname name   Hostname field (default: this machine)
  --appname name    App-name field (default: sendlog)
  --format f        rfc5424 (default), rfc3164 or raw-pri

Output:
  --json            Emit a single JSON result line on stdout
`

var (
	version = "dev"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "send":
		err = cmdSend(args)
	case "connect":
		err = cmdConnect(args)
	case "stress":
		err = cmdStress(args)
	case "profile":
		err = cmdProfile(args)
	case "template":
		err = cmdTemplate(args)
	case "import":
		err = cmdImport(args)
	case "export":
		err = cmdExport()
	case "version", "--version", "-v":
		fmt.Println(version)
		return
	case "help", "--help", "-h":
		fmt.Fprint(os.Stderr, usageText)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usageText)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// ================================================================================
// Shared flag plumbing
// ================================================================================

// connFlags holds the connection-related options shared by send/connect/stress
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
	fs.StringVar(&c.address, "address", "", "host or IP (alias)")
	fs.StringVar(&c.port, "p", "", "port")
	fs.StringVar(&c.port, "port", "", "port (alias)")
	fs.StringVar(&c.protocol, "t", "tcp", "protocol: tcp or udp")
	fs.StringVar(&c.protocol, "protocol", "tcp", "protocol (alias)")
	fs.StringVar(&c.framing, "framing", "", "tcp framing: octet-counting or non-transparent")
	fs.BoolVar(&c.useTLS, "tls", false, "enable TLS (tcp only)")
	fs.BoolVar(&c.tlsVerify, "verify", true, "verify TLS certificates")
	fs.BoolVar(&c.json, "json", false, "print a single JSON result line to stdout")
	fs.StringVar(&c.ca, "ca", "", "CA certificate path (PEM)")
	fs.StringVar(&c.cert, "cert", "", "client certificate path (PEM)")
	fs.StringVar(&c.key, "key", "", "client key path (PEM)")
}

// msgFlags holds the message-formatting options
type msgFlags struct {
	messages multiFlag
	file     string
	facility int
	severity int
	hostname string
	appname  string
	format   string
	use5424  bool
}

func (m *msgFlags) register(fs *flag.FlagSet) {
	fs.Var(&m.messages, "m", "message to send (repeatable)")
	fs.Var(&m.messages, "message", "message to send (repeatable, alias)")
	fs.StringVar(&m.file, "f", "", "log file to import (.json/.csv/.txt)")
	fs.StringVar(&m.file, "file", "", "log file to import (alias)")
	fs.IntVar(&m.facility, "facility", 16, "facility 0-23")
	fs.IntVar(&m.severity, "severity", 6, "severity 0-7")
	fs.StringVar(&m.hostname, "hostname", "", "hostname field (default: this machine)")
	fs.StringVar(&m.appname, "appname", "sendlog", "app-name field")
	fs.StringVar(&m.format, "format", "rfc5424", "message format: rfc5424, rfc3164 or raw-pri")
	fs.BoolVar(&m.use5424, "rfc5424", true, "legacy: use RFC 5424 when --format is empty")
}

func (m *msgFlags) intoConfig() syslog.MessageFormat {
	switch m.format {
	case "rfc5424":
		return syslog.MessageFormatRFC5424
	case "rfc3164":
		return syslog.MessageFormatRFC3164
	case "raw-pri":
		return syslog.MessageFormatRawPRI
	case "":
		return "" // falls back to -rfc5424 legacy switch
	default:
		return syslog.MessageFormat(m.format) // invalid value surfaces from ValidateConfig
	}
}

// multiFlag is a repeatable string flag
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func dialConfig(c *connFlags, m *msgFlags) syslog.Config {
	return syslog.Config{
		Address:        c.address,
		Port:           c.port,
		Protocol:       strings.ToLower(c.protocol),
		Method:         framing.Method(c.framing),
		Facility:       uint8(m.facility),
		Severity:       uint8(m.severity),
		Hostname:       m.hostname,
		Appname:        m.appname,
		MessageFormat:  m.intoConfig(),
		UseRFC5424:     m.use5424,
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

// ================================================================================
// send
// ================================================================================

type sendResult struct {
	OK     bool     `json:"ok"`
	Sent   int      `json:"sent"`
	Errors []string `json:"errors,omitempty"`
}

func cmdSend(args []string) error {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	var c connFlags
	var m msgFlags
	c.register(fs)
	m.register(fs)
	_ = fs.Parse(args)

	config := dialConfig(&c, &m)

	if m.file != "" {
		result, err := logimport.Import(m.file)
		if err != nil {
			return err
		}
		config.Messages = append(config.Messages, result.Messages...)
		// Import errors are reported but do not block sending valid messages
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "import warning: %s\n", e)
		}
	}
	config.Messages = append(config.Messages, m.messages...)

	if len(config.Messages) == 0 {
		return fmt.Errorf("no messages to send (use -m or -f)")
	}

	resp := syslog.SendSyslogMessages(config, func(current, total int, percent float64) {
		if !c.json {
			fmt.Fprintf(os.Stderr, "\rprogress: %d/%d (%.0f%%)", current, total, percent)
		}
	})
	if !c.json {
		fmt.Fprintln(os.Stderr)
	}

	ok := len(resp.Errors) == 0 && len(resp.SentMessages) > 0
	if c.json {
		errs := resp.Errors
		if errs == nil {
			errs = []string{}
		}
		emitJSON(sendResult{OK: ok, Sent: len(resp.SentMessages), Errors: errs})
	} else {
		fmt.Printf("sent %d message(s), %d error(s)\n", len(resp.SentMessages), len(resp.Errors))
		for _, e := range resp.Errors {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
	}

	if !ok {
		return fmt.Errorf("send finished with errors")
	}
	return nil
}

// ================================================================================
// connect
// ================================================================================

func cmdConnect(args []string) error {
	fs := flag.NewFlagSet("connect", flag.ExitOnError)
	var c connFlags
	c.register(fs)
	_ = fs.Parse(args)

	protocolInfo, err := syslog.CheckConnection(
		c.address, c.port, strings.ToLower(c.protocol),
		c.useTLS, c.tlsVerify, c.ca, c.cert, c.key,
	)
	if err != nil {
		if c.json {
			emitJSON(map[string]interface{}{"ok": false, "error": err.Error()})
		}
		return err
	}

	if c.json {
		emitJSON(map[string]interface{}{"ok": true, "protocol": protocolInfo})
	} else {
		fmt.Printf("connection ok (%s) to %s:%s\n", protocolInfo, c.address, c.port)
	}
	return nil
}

// ================================================================================
// stress
// ================================================================================

func cmdStress(args []string) error {
	fs := flag.NewFlagSet("stress", flag.ExitOnError)
	var c connFlags
	var m msgFlags
	c.register(fs)
	m.register(fs)
	rate := fs.Int("rate", 10, "messages per second")
	duration := fs.Int("duration", 0, "stop after N seconds (0 = until Ctrl+C)")
	maxMessages := fs.Int("max-messages", 0, "stop after N messages (0 = unlimited)")
	randomize := fs.Bool("randomize", false, "prefix each message with sequence and timestamp")
	_ = fs.Parse(args)

	// Base message: first -m/-message, else the default string
	base := "stress test"
	if len(m.messages) > 0 {
		base = m.messages[0]
	}

	engine := stress.NewEngine()
	defer engine.Stop()

	config := stress.ContinuousSendConfig{
		Address:        c.address,
		Port:           c.port,
		Protocol:       strings.ToLower(c.protocol),
		Message:        base,
		Method:         framing.Method(c.framing),
		Facility:       uint8(m.facility),
		Severity:       uint8(m.severity),
		Hostname:       m.hostname,
		Appname:        m.appname,
		MessageFormat:  m.intoConfig(),
		UseRFC5424:     m.use5424,
		UseTLS:         c.useTLS,
		TLSVerify:      c.tlsVerify,
		CACertPath:     c.ca,
		ClientCertPath: c.cert,
		ClientKeyPath:  c.key,
		Duration:       *duration,
		MessagesPerSec: *rate,
		MaxMessages:    *maxMessages,
		RandomizeData:  *randomize,
	}

	// Graceful stop on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nstopping...")
		engine.Stop()
	}()

	lastPrint := time.Time{}
	err := engine.Start(config, func(stats stress.ContinuousStats) {
		if stats.IsRunning && time.Since(lastPrint) >= 500*time.Millisecond {
			lastPrint = time.Now()
			fmt.Fprintf(os.Stderr, "\r[%.1fs] sent=%d err=%d rate=%.0f/s",
				stats.ElapsedSeconds, stats.TotalSent, stats.TotalErrors, stats.CurrentRate)
		}
	})
	if err != nil {
		return err
	}

	// Wait until the engine stops on its own (duration / max / signal)
	for engine.IsContinuousRunning() {
		time.Sleep(100 * time.Millisecond)
	}

	stats := engine.GetContinuousStats()
	if c.json {
		emitJSON(stats)
	} else {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Printf("stopped after %.1fs: sent=%d errors=%d avgRate=%.0f/s\n",
			stats.ElapsedSeconds, stats.TotalSent, stats.TotalErrors, stats.CurrentRate)
	}
	return nil
}

// ================================================================================
// profile / template / import / export
// ================================================================================

type profileJSON struct {
	OK       bool                        `json:"ok"`
	Profiles []storage.ConnectionProfile `json:"profiles,omitempty"`
	Error    string                      `json:"error,omitempty"`
}

func cmdProfile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sendlog-cli profile list|add|remove <id>")
	}
	switch args[0] {
	case "list":
		return profileList(args[1:])
	case "add":
		return profileAdd(args[1:])
	case "remove":
		return profileRemove(args[1:])
	default:
		return fmt.Errorf("unknown profile subcommand %q (want list|add|remove)", args[0])
	}
}

func profileList(args []string) error {
	fs := flag.NewFlagSet("profile list", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)

	profiles, err := storage.GetProfiles()
	if err != nil {
		return err
	}
	if *jsonOut {
		if profiles == nil {
			profiles = []storage.ConnectionProfile{}
		}
		emitJSON(profileJSON{OK: true, Profiles: profiles})
		return nil
	}
	if len(profiles) == 0 {
		fmt.Println("no saved profiles")
		return nil
	}
	for _, p := range profiles {
		proto := p.Protocol
		if p.UseTLS {
			proto += "+TLS"
		}
		fmt.Printf("%s  %s:%s  %s  framing=%s\n", p.ID, p.Address, p.Port, proto, p.Method)
		if p.Name != "" {
			fmt.Printf("      name=%s\n", p.Name)
		}
	}
	return nil
}

func profileAdd(args []string) error {
	fs := flag.NewFlagSet("profile add", flag.ExitOnError)
	name := fs.String("name", "", "profile name (required)")
	desc := fs.String("description", "", "description")
	c := connFlags{}
	c.register(fs)
	_ = fs.Parse(args)

	if *name == "" {
		return fmt.Errorf("--name is required")
	}

	profile := storage.ConnectionProfile{
		Name:           *name,
		Description:    *desc,
		Address:        c.address,
		Port:           c.port,
		Protocol:       strings.ToLower(c.protocol),
		Method:         c.framing,
		UseTLS:         c.useTLS,
		TLSVerify:      c.tlsVerify,
		CACertPath:     c.ca,
		ClientCertPath: c.cert,
		ClientKeyPath:  c.key,
	}
	saved, err := storage.SaveProfile(profile)
	if err != nil {
		return err
	}
	fmt.Printf("saved profile %s (%s)\n", saved.Name, saved.ID)
	return nil
}

func profileRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sendlog-cli profile remove <id>")
	}
	if err := storage.DeleteProfile(args[0]); err != nil {
		return err
	}
	fmt.Printf("deleted profile %s\n", args[0])
	return nil
}

func cmdTemplate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: sendlog-cli template list")
	}
	if args[0] != "list" {
		return fmt.Errorf("unknown template subcommand %q (want list)", args[0])
	}
	fs := flag.NewFlagSet("template list", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args[1:])

	templates, err := storage.GetTemplates()
	if err != nil {
		return err
	}
	if *jsonOut {
		if templates == nil {
			templates = []storage.LogTemplate{}
		}
		emitJSON(map[string]interface{}{"ok": true, "templates": templates})
		return nil
	}
	if len(templates) == 0 {
		fmt.Println("no saved templates")
		return nil
	}
	for _, t := range templates {
		fmt.Printf("%s  %s  facility=%d severity=%d\n", t.ID, t.Name, t.Facility, t.Severity)
	}
	return nil
}

func cmdImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	file := fs.String("f", "", "JSON file to import (required)")
	fs.StringVar(file, "file", "", "JSON file to import (alias)")
	merge := fs.Bool("merge", true, "merge with existing storage (default) vs replace")
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)

	if *file == "" {
		return fmt.Errorf("-f <file> is required")
	}
	data, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	if err := storage.Import(string(data), *merge); err != nil {
		if *jsonOut {
			emitJSON(map[string]interface{}{"ok": false, "error": err.Error()})
		}
		return err
	}
	mode := "merged"
	if !*merge {
		mode = "replaced"
	}
	if *jsonOut {
		emitJSON(map[string]interface{}{"ok": true, "mode": mode})
	} else {
		fmt.Printf("import %s\n", mode)
	}
	return nil
}

func cmdExport() error {
	data, err := storage.Export()
	if err != nil {
		return err
	}
	fmt.Println(data)
	return nil
}
