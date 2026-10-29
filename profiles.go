package main

import (
	"context"

	"Sendlog-Syslog/internal/storage"
)

// ================================================================================
// PROFILE SERVICE - Manages connection profiles (Enterprise Feature)
// Follows Single Responsibility Principle: handles only profile CRUD operations
// The persistence itself lives in internal/storage, shared with the CLI.
// ================================================================================

// ConnectionProfile aliases the internal profile type so the frontend binding shape is unchanged
type ConnectionProfile = storage.ConnectionProfile

// ProfileService handles connection profile operations
type ProfileService struct {
	ctx context.Context
}

// NewProfileService creates a new ProfileService instance
func NewProfileService() *ProfileService {
	return &ProfileService{}
}

// SetContext sets the Wails runtime context (kept for symmetry with other services)
func (p *ProfileService) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// GetProfiles returns all saved connection profiles
func (p *ProfileService) GetProfiles() ([]ConnectionProfile, error) {
	return storage.GetProfiles()
}

// SaveProfile saves a new or updates an existing connection profile
func (p *ProfileService) SaveProfile(profile ConnectionProfile) (ConnectionProfile, error) {
	return storage.SaveProfile(profile)
}

// DeleteProfile deletes a connection profile by ID
func (p *ProfileService) DeleteProfile(id string) error {
	return storage.DeleteProfile(id)
}
