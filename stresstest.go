package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Sendlog-Syslog/internal/stress"
)

// ================================================================================
// STRESS TEST SERVICE - GUI-facing wrapper around the internal/stress engine
// The continuous-send loop lives in internal/stress; this service wires the
// Wails runtime context and maps stats into runtime events.
// ================================================================================

// ContinuousSendConfig aliases the internal config so the frontend binding shape is unchanged
type ContinuousSendConfig = stress.ContinuousSendConfig

// ContinuousStats aliases the internal stats
type ContinuousStats = stress.ContinuousStats

// StressTestService handles continuous/stress test operations for the GUI
type StressTestService struct {
	ctx    context.Context
	engine *stress.Engine
}

// NewStressTestService creates a new StressTestService instance
func NewStressTestService() *StressTestService {
	return &StressTestService{
		engine: stress.NewEngine(),
	}
}

// SetContext sets the Wails runtime context
func (s *StressTestService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// IsContinuousRunning returns whether continuous send is currently active
func (s *StressTestService) IsContinuousRunning() bool {
	return s.engine.IsContinuousRunning()
}

// GetContinuousStats returns current statistics for continuous send
func (s *StressTestService) GetContinuousStats() ContinuousStats {
	return s.engine.GetContinuousStats()
}

// StartContinuousSend starts sending messages continuously based on config
func (s *StressTestService) StartContinuousSend(config ContinuousSendConfig) error {
	if s.ctx == nil {
		return fmt.Errorf("stress test service has no Wails context (SetContext not called)")
	}
	return s.engine.Start(config, func(stats ContinuousStats) {
		runtime.EventsEmit(s.ctx, "continuous:stats", stats)
	})
}

// StopContinuousSend stops the continuous send operation
func (s *StressTestService) StopContinuousSend() {
	if s.ctx == nil {
		return
	}
	s.engine.Stop()
	runtime.EventsEmit(s.ctx, "continuous:stopped", s.engine.GetContinuousStats())
}
