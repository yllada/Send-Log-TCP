package main

import (
	"context"
)

// ================================================================================
// PROFILE SERVICE - Manages connection profiles (Enterprise Feature)
// Follows Single Responsibility Principle: handles only profile CRUD operations
// ================================================================================

// ConnectionProfile represents a saved connection configuration
type ConnectionProfile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Address        string `json:"address"`
	Port           string `json:"port"`
	Protocol       string `json:"protocol"`
	FramingMethod  string `json:"framingMethod"`
	UseTLS         bool   `json:"useTls"`
	TLSVerify      bool   `json:"tlsVerify"`
	CACertPath     string `json:"caCertPath,omitempty"`
	ClientCertPath string `json:"clientCertPath,omitempty"`
	ClientKeyPath  string `json:"clientKeyPath,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// ProfileService handles connection profile operations
type ProfileService struct {
	ctx context.Context
}

// NewProfileService creates a new ProfileService instance
func NewProfileService() *ProfileService {
	return &ProfileService{}
}

// SetContext sets the Wails runtime context
func (p *ProfileService) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// GetProfiles returns all saved connection profiles
func (p *ProfileService) GetProfiles() ([]ConnectionProfile, error) {
	return getProfilesFromStorage()
}

// SaveProfile saves a new or updates an existing connection profile
func (p *ProfileService) SaveProfile(profile ConnectionProfile) (ConnectionProfile, error) {
	return saveProfileToStorage(profile)
}

// DeleteProfile deletes a connection profile by ID
func (p *ProfileService) DeleteProfile(id string) error {
	return deleteProfileFromStorage(id)
}
