package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Sendlog-Syslog/internal/logimport"
)

// ================================================================================
// BATCH IMPORT SERVICE - GUI-facing wrapper around internal/logimport
// The file parsing itself lives in internal/logimport; this service only maps
// the Wails file dialog onto it.
// ================================================================================

// BatchImportResult aliases the internal result so the frontend binding shape is unchanged
type BatchImportResult = logimport.Result

// BatchImportService handles batch log import operations for the GUI
type BatchImportService struct {
	ctx context.Context
}

// NewBatchImportService creates a new BatchImportService instance
func NewBatchImportService() *BatchImportService {
	return &BatchImportService{}
}

// SetContext sets the Wails runtime context
func (b *BatchImportService) SetContext(ctx context.Context) {
	b.ctx = ctx
}

// SelectLogFile opens a file dialog to select a log file for batch import
func (b *BatchImportService) SelectLogFile() (string, error) {
	filePath, err := runtime.OpenFileDialog(b.ctx, runtime.OpenDialogOptions{
		Title: "Select Log File for Import",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Log Files (*.csv, *.json, *.txt, *.log)",
				Pattern:     "*.csv;*.json;*.txt;*.log",
			},
			{
				DisplayName: "CSV Files (*.csv)",
				Pattern:     "*.csv",
			},
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "Text Files (*.txt, *.log)",
				Pattern:     "*.txt;*.log",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// ImportLogFile reads and parses a log file, returning the messages
func (b *BatchImportService) ImportLogFile(filePath string) (BatchImportResult, error) {
	return logimport.Import(filePath)
}
