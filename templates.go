package main

import (
	"context"

	"Sendlog-Syslog/internal/storage"
)

// ================================================================================
// TEMPLATE SERVICE - Manages log templates (Enterprise Feature)
// Follows Single Responsibility Principle: handles only template CRUD operations
// The persistence itself lives in internal/storage, shared with the CLI.
// ================================================================================

// LogTemplate aliases the internal template type so the frontend binding shape is unchanged
type LogTemplate = storage.LogTemplate

// TemplateService handles log template operations
type TemplateService struct {
	ctx context.Context
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// SetContext sets the Wails runtime context (kept for symmetry with other services)
func (t *TemplateService) SetContext(ctx context.Context) {
	t.ctx = ctx
}

// GetTemplates returns all saved log templates
func (t *TemplateService) GetTemplates() ([]LogTemplate, error) {
	return storage.GetTemplates()
}

// SaveTemplate saves a new or updates an existing log template
func (t *TemplateService) SaveTemplate(template LogTemplate) (LogTemplate, error) {
	return storage.SaveTemplate(template)
}

// DeleteTemplate deletes a log template by ID
func (t *TemplateService) DeleteTemplate(id string) error {
	return storage.DeleteTemplate(id)
}

// ================================================================================
// CONFIG IMPORT/EXPORT (Enterprise Feature)
// ================================================================================

// ExportConfig exports all profiles and templates to a JSON string
func (t *TemplateService) ExportConfig() (string, error) {
	return storage.Export()
}

// ImportConfig imports profiles and templates from a JSON string
func (t *TemplateService) ImportConfig(jsonData string, merge bool) error {
	return storage.Import(jsonData, merge)
}
