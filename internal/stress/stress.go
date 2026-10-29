// Package stress implements the continuous / stress-test syslog send engine.
// It is independent of the Wails GUI; stats are exposed via callbacks so the
// GUI can emit runtime events and the CLI can print to stderr.
package stress

import (
	"context"
	"fmt"
	"sync"
	"time"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/syslog"
)

// ContinuousSendConfig configuration for continuous/stress test mode
type ContinuousSendConfig struct {
	// Connection settings
	Address       string        `json:"Address"`
	Port          string        `json:"Port"`
	Protocol      string        `json:"Protocol"`
	Message       string        `json:"Message"`
	FramingMethod framing.FramingMethod `json:"FramingMethod"`
	Facility      uint8         `json:"Facility"`
	Severity      uint8         `json:"Severity"`
	Hostname      string        `json:"Hostname"`
	Appname       string        `json:"Appname"`
	MessageFormat syslog.MessageFormat `json:"MessageFormat"`
	// UseRFC5424 is the legacy format switch, kept for stored configs.
	// MessageFormat takes precedence when set.
	UseRFC5424     bool   `json:"UseRFC5424"`
	UseTLS         bool   `json:"UseTLS"`
	TLSVerify      bool   `json:"TLSVerify"`
	CACertPath     string `json:"CACertPath"`
	ClientCertPath string `json:"ClientCertPath"`
	ClientKeyPath  string `json:"ClientKeyPath"`

	// Continuous send settings
	Duration       int  `json:"Duration"`
	MessagesPerSec int  `json:"MessagesPerSec"`
	MaxMessages    int  `json:"MaxMessages"`
	RandomizeData  bool `json:"RandomizeData"`
}

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
// onStats is called (may be nil) on the ~250ms stats tick and on stop.
func (s *Engine) Start(config ContinuousSendConfig, onStats func(ContinuousStats)) error {
	s.continuousMu.Lock()
	if s.isRunning {
		s.continuousMu.Unlock()
		return fmt.Errorf("continuous send already running")
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

// Stop stops the continuous send operation
func (s *Engine) Stop() {
	s.continuousMu.Lock()
	defer s.continuousMu.Unlock()

	if s.continuousCancel != nil {
		s.continuousCancel()
		s.continuousCancel = nil
	}
	s.isRunning = false
	s.continuousStats.IsRunning = false
}

// runContinuousSend is the goroutine that handles continuous message sending
func (s *Engine) runContinuousSend(ctx context.Context, config ContinuousSendConfig, onStats func(ContinuousStats)) {
	defer func() {
		s.continuousMu.Lock()
		s.isRunning = false
		s.continuousStats.IsRunning = false
		stats := s.continuousStats
		s.continuousMu.Unlock()
		if onStats != nil {
			onStats(stats)
		}
	}()

	// Build syslog config
	syslogConfig := syslog.SyslogConfig{
		Address:        config.Address,
		Port:           config.Port,
		Protocol:       config.Protocol,
		FramingMethod:  config.FramingMethod,
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

	// Validate config
	if err := syslog.ValidateConfig(&syslogConfig); err != nil {
		return
	}

	// Establish connection
	conn, err := syslog.Dial(syslogConfig)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	// Create framer for TCP
	var framer *framing.Framer
	if config.Protocol == "tcp" {
		framer = framing.NewFramer(framing.FramingConfig{
			Method:           config.FramingMethod,
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

	// Main loop
	for {
		select {
		case <-ctx.Done():
			return

		case <-durationTimer:
			return

		case <-ticker.C:
			// Check max messages limit
			if config.MaxMessages > 0 && sentCount >= int64(config.MaxMessages) {
				return
			}

			// Build message with optional randomization
			message := config.Message
			if config.RandomizeData {
				sequenceNum++
				message = fmt.Sprintf("[seq=%d ts=%s] %s",
					sequenceNum,
					time.Now().Format("15:04:05.000"),
					config.Message,
				)
			}

			// Build syslog message
			syslogMsg, err := syslog.BuildMessage(syslogConfig, message)
			if err != nil {
				errorCount++
				continue
			}

			// Send message
			var sendErr error
			if config.Protocol == "tcp" && framer != nil {
				framedMsg, err := framer.Frame(syslogMsg)
				if err != nil {
					errorCount++
					continue
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

			// Update stats every 250ms
			if time.Since(lastStatsUpdate) >= 250*time.Millisecond {
				elapsed := time.Since(startTime).Seconds()
				currentRate := float64(sentCount) / elapsed

				s.continuousMu.Lock()
				s.continuousStats.TotalSent = sentCount
				s.continuousStats.TotalErrors = errorCount
				s.continuousStats.ElapsedSeconds = elapsed
				s.continuousStats.CurrentRate = currentRate
				stats := s.continuousStats
				s.continuousMu.Unlock()

				if onStats != nil {
					onStats(stats)
				}
				lastStatsUpdate = time.Now()
			}
		}
	}
}
