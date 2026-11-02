// Package stress implements the continuous send engine used for load testing the syslog connection.
package stress

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/syslog"
)

// ContinuousSendConfig configures continuous sending
type ContinuousSendConfig struct {
	Address        string               `json:"address"`
	Port           string               `json:"port"`
	Protocol       string               `json:"protocol"`
	Method         framing.Method       `json:"method"`
	Facility       uint8                `json:"facility"`
	Severity       uint8                `json:"severity"`
	Hostname       string               `json:"hostname"`
	Appname        string               `json:"appname"`
	MessageFormat  syslog.MessageFormat `json:"messageFormat"`
	UseRFC5424     bool                 `json:"useRFC5424"`
	Message        string               `json:"message"`
	MessagesPerSec int                  `json:"messagesPerSec"`
	Duration       int                  `json:"duration"`
	MaxMessages    int                  `json:"maxMessages"`
	RandomizeData  bool                 `json:"randomizeData"`
	UseTLS         bool                 `json:"useTLS"`
	TLSVerify      bool                 `json:"tlsVerify"`
	CACertPath     string               `json:"caCertPath"`
	ClientCertPath string               `json:"clientCertPath"`
	ClientKeyPath  string               `json:"clientKeyPath"`
	// Conn is an already-open transport the engine sends over without closing
	// it, for reuse with the GUI's live connection.
	Conn net.Conn `json:"-"`
	// ownsConn is true when the engine opened Conn itself and must close it.
	ownsConn bool
}

// OwnsConnection reports whether the engine opened Conn and must close it
// when the run ends; a caller-supplied Conn stays open.
func (c ContinuousSendConfig) OwnsConnection() bool { return c.ownsConn }

// ContinuousStats real-time statistics for continuous send
type ContinuousStats struct {
	TotalSent      int64   `json:"totalSent"`
	TotalErrors    int64   `json:"totalErrors"`
	CurrentRate    float64 `json:"currentRate"`
	ElapsedSeconds float64 `json:"elapsedSeconds"`
	IsRunning      bool    `json:"isRunning"`
	StartTime      int64   `json:"startTime"`
	TargetRate     int     `json:"targetRate"`
	Duration       int     `json:"duration"`
}

// Engine handles continuous/stress test operations
type Engine struct {
	continuousCancel context.CancelFunc
	continuousMu     sync.RWMutex
	continuousStats  ContinuousStats
	isRunning        bool
}

// NewEngine creates a new Engine instance
func NewEngine() *Engine {
	return &Engine{}
}

// IsContinuousRunning returns whether continuous send is currently active
func (s *Engine) IsContinuousRunning() bool {
	s.continuousMu.RLock()
	defer s.continuousMu.RUnlock()
	return s.isRunning
}

// GetContinuousStats returns current statistics for continuous send
func (s *Engine) GetContinuousStats() ContinuousStats {
	s.continuousMu.RLock()
	defer s.continuousMu.RUnlock()
	return s.continuousStats
}

// Start begins sending messages continuously based on config.
// It validates the configuration and (when config.Conn is nil) dials the
// target before returning, so a bad setup never leaves the engine running.
// onStats is called (may be nil) on the ~250ms stats tick and on stop.
func (s *Engine) Start(config ContinuousSendConfig, onStats func(ContinuousStats)) error {
	s.continuousMu.Lock()
	if s.isRunning {
		s.continuousMu.Unlock()
		return fmt.Errorf("continuous send already running")
	}

	// Build syslog config for validation
	syslogConfig := syslog.Config{
		Address:        config.Address,
		Port:           config.Port,
		Protocol:       config.Protocol,
		Method:         config.Method,
		Facility:       config.Facility,
		Severity:       config.Severity,
		Hostname:       config.Hostname,
		Appname:        config.Appname,
		MessageFormat:  config.MessageFormat,
		UseRFC5424:     config.UseRFC5424,
		UseTLS:         config.UseTLS,
		TLSVerify:      config.TLSVerify,
		CACertPath:     config.CACertPath,
		ClientCertPath: config.ClientCertPath,
		ClientKeyPath:  config.ClientKeyPath,
	}
	if err := syslog.ValidateConfig(&syslogConfig); err != nil {
		s.continuousMu.Unlock()
		return err
	}
	if config.Conn == nil {
		conn, err := syslog.Dial(syslogConfig)
		if err != nil {
			s.continuousMu.Unlock()
			return err
		}
		config.Conn = conn
		config.ownsConn = true
	}

	// Create cancellation context
	ctx, cancel := context.WithCancel(context.Background())
	s.continuousCancel = cancel
	s.isRunning = true
	s.continuousStats = ContinuousStats{
		TotalSent:   0,
		TotalErrors: 0,
		CurrentRate: 0,
		IsRunning:   true,
		StartTime:   time.Now().Unix(),
		TargetRate:  config.MessagesPerSec,
		Duration:    config.Duration,
	}
	s.continuousMu.Unlock()

	// Start the background goroutine
	go s.runContinuousSend(ctx, config, onStats)

	return nil
}

// Stop stops the continuous sending
func (s *Engine) Stop() {
	s.continuousMu.Lock()
	if s.continuousCancel != nil {
		s.continuousCancel()
	}
	s.continuousMu.Unlock()
}

// runContinuousSend is the goroutine that handles continuous message sending
func (s *Engine) runContinuousSend(ctx context.Context, config ContinuousSendConfig, onStats func(ContinuousStats)) {
	// Build syslog config
	syslogConfig := syslog.Config{
		Address:        config.Address,
		Port:           config.Port,
		Protocol:       config.Protocol,
		Method:         config.Method,
		Facility:       config.Facility,
		Severity:       config.Severity,
		Hostname:       config.Hostname,
		Appname:        config.Appname,
		MessageFormat:  config.MessageFormat,
		UseRFC5424:     config.UseRFC5424,
		UseTLS:         config.UseTLS,
		TLSVerify:      config.TLSVerify,
		CACertPath:     config.CACertPath,
		ClientCertPath: config.ClientCertPath,
		ClientKeyPath:  config.ClientKeyPath,
	}

	conn := config.Conn

	// Create framer for TCP
	var framer *framing.Framer
	if config.Protocol == "tcp" {
		framer = framing.NewFramer(framing.Config{
			Method:           config.Method,
			ValidateUTF8:     true,
			MaxMessageLength: 0,
		})
	}

	// Rate limiting with ticker
	interval := time.Second / time.Duration(config.MessagesPerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Duration timer if configured
	var durationTimer <-chan time.Time
	if config.Duration > 0 {
		durationTimer = time.After(time.Duration(config.Duration) * time.Second)
	}

	// Stats tracking
	startTime := time.Now()
	var sentCount, errorCount int64
	var sequenceNum int64
	lastStatsUpdate := time.Now()

	publish := func(running bool) {
		elapsed := time.Since(startTime).Seconds()
		currentRate := 0.0
		if elapsed > 0 {
			currentRate = float64(sentCount) / elapsed
		}

		s.continuousMu.Lock()
		s.isRunning = running
		s.continuousStats.TotalSent = sentCount
		s.continuousStats.TotalErrors = errorCount
		s.continuousStats.ElapsedSeconds = elapsed
		s.continuousStats.CurrentRate = currentRate
		s.continuousStats.IsRunning = running
		stats := s.continuousStats
		s.continuousMu.Unlock()

		if onStats != nil {
			onStats(stats)
		}
	}

	done := false
	for !done {
		select {
		case <-ctx.Done():
			done = true
		case <-durationTimer:
			done = true
		case <-ticker.C:
			if config.MaxMessages > 0 && sentCount >= int64(config.MaxMessages) {
				done = true
				break
			}

			message := config.Message
			if config.RandomizeData {
				sequenceNum++
				message = fmt.Sprintf("[seq=%d ts=%s] %s",
					sequenceNum,
					time.Now().Format("15:04:05.000"),
					config.Message,
				)
			}

			syslogMsg, err := syslog.BuildMessage(syslogConfig, message)
			if err != nil {
				errorCount++
				break
			}

			var sendErr error
			if config.Protocol == "tcp" && framer != nil {
				framedMsg, err := framer.Frame(syslogMsg)
				if err != nil {
					errorCount++
					break
				}
				if err := conn.SetWriteDeadline(time.Now().Add(syslog.WriteTimeout)); err != nil {
					_ = err
				}
				sendErr = syslog.WriteAll(conn, framedMsg)
			} else {
				if err := conn.SetWriteDeadline(time.Now().Add(syslog.WriteTimeout)); err != nil {
					_ = err
				}
				sendErr = syslog.WriteAll(conn, []byte(syslogMsg))
			}

			if sendErr != nil {
				errorCount++
			} else {
				sentCount++
			}

			if time.Since(lastStatsUpdate) >= 250*time.Millisecond {
				publish(true)
				lastStatsUpdate = time.Now()
			}
		}
	}

	// Final stats flush so short runs report their true totals.
	publish(false)

	if config.OwnsConnection() {
		_ = conn.Close()
	}
}
