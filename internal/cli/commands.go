package cli

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/logimport"
	"Sendlog-Syslog/internal/stress"
	"Sendlog-Syslog/internal/syslog"
)

type sendResult struct {
	OK     bool     `json:"ok"`
	Sent   int      `json:"sent"`
	Errors []string `json:"errors,omitempty"`
}

// RunSend sends the flagged messages and reports the result.
func RunSend(args []string) error {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	var c connFlags
	var m msgFlags
	c.register(fs)
	m.register(fs)
	_ = fs.Parse(args)

	config, err := connConfig(&c, &m)
	if err != nil {
		return err
	}

	if m.file != "" {
		result, err := logimport.Import(m.file)
		if err != nil {
			return err
		}
		config.Messages = append(config.Messages, result.Messages...)
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

// RunConnect tests a syslog connection and reports the negotiated protocol.
func RunConnect(args []string) error {
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

// RunStress starts a continuous send and blocks until stopped.
func RunStress(args []string) error {
	fs := flag.NewFlagSet("stress", flag.ExitOnError)
	var c connFlags
	var m msgFlags
	c.register(fs)
	m.register(fs)
	rate := fs.Int("rate", 10, "messages per second")
	duration := fs.Int("duration", 0, "stop after N seconds (0 = until stopped)")
	maxMessages := fs.Int("max", 0, "stop after N messages (0 = unlimited)")
	_ = fs.Parse(args)

	base := "stress test"
	if len(m.messages) > 0 {
		base = m.messages[0]
	}

	facility, severity, err := facilitySeverity(&m)
	if err != nil {
		return err
	}

	config := stress.ContinuousSendConfig{
		Address:        c.address,
		Port:           c.port,
		Protocol:       strings.ToLower(c.protocol),
		Message:        base,
		Method:         framing.Method(c.framing),
		Facility:       facility,
		Severity:       severity,
		Hostname:       m.hostname,
		Appname:        m.appname,
		MessageFormat:  m.messageFormat(),
		UseTLS:         c.useTLS,
		TLSVerify:      c.tlsVerify,
		CACertPath:     c.ca,
		ClientCertPath: c.cert,
		ClientKeyPath:  c.key,
		Duration:       *duration,
		MessagesPerSec: *rate,
		MaxMessages:    *maxMessages,
	}
	return runStress(config, c.json)
}

func runStress(config stress.ContinuousSendConfig, jsonOut bool) error {
	engine := stress.NewEngine()
	defer engine.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nstopping...")
		engine.Stop()
	}()

	lastPrint := time.Time{}
	if err := engine.Start(config, func(stats stress.ContinuousStats) {
		if stats.IsRunning && time.Since(lastPrint) >= 500*time.Millisecond {
			lastPrint = time.Now()
			fmt.Fprintf(os.Stderr, "\r[%.1fs] sent=%d err=%d rate=%.0f/s",
				stats.ElapsedSeconds, stats.TotalSent, stats.TotalErrors, stats.CurrentRate)
		}
	}); err != nil {
		return err
	}

	for engine.IsContinuousRunning() {
		time.Sleep(100 * time.Millisecond)
	}

	stats := engine.GetContinuousStats()
	if jsonOut {
		emitJSON(stats)
	} else {
		fmt.Fprintln(os.Stderr)
		fmt.Printf("stopped after %.1fs: sent=%d errors=%d avgRate=%.0f/s\n",
			stats.ElapsedSeconds, stats.TotalSent, stats.TotalErrors, stats.CurrentRate)
	}
	return nil
}
