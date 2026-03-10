package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ================================================================================
// CONNECTION PROFILES & LOG TEMPLATES STORAGE
// Enterprise feature: Persistent storage for user configurations
// ================================================================================

// ConnectionProfile represents a saved connection configuration
type ConnectionProfile struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description,omitempty"`
	Address        string        `json:"address"`
	Port           string        `json:"port"`
	Protocol       string        `json:"protocol"`
	FramingMethod  FramingMethod `json:"framingMethod"`
	UseTLS         bool          `json:"useTls"`
	TLSVerify      bool          `json:"tlsVerify"`
	CACertPath     string        `json:"caCertPath,omitempty"`
	ClientCertPath string        `json:"clientCertPath,omitempty"`
	ClientKeyPath  string        `json:"clientKeyPath,omitempty"`
	CreatedAt      int64         `json:"createdAt"`
	UpdatedAt      int64         `json:"updatedAt"`
}

// LogTemplate represents a reusable log message template
type LogTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Message     string `json:"message"`
	Facility    uint8  `json:"facility"`
	Severity    uint8  `json:"severity"`
	Appname     string `json:"appname"`
	UseRFC5424  bool   `json:"useRfc5424"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// StorageData represents the complete storage file structure
type StorageData struct {
	Version   int                 `json:"version"`
	Profiles  []ConnectionProfile `json:"profiles"`
	Templates []LogTemplate       `json:"templates"`
}

const (
	storageVersion  = 1
	storageFileName = "sendlog-config.json"
	appFolderName   = "SendLog-Syslog"
)

// getStoragePath returns the path to the storage file
// Uses XDG config on Linux, AppData on Windows, Library on macOS
func getStoragePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	appDir := filepath.Join(configDir, appFolderName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(appDir, storageFileName), nil
}

// loadStorage loads the storage data from disk
func loadStorage() (*StorageData, error) {
	path, err := getStoragePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty storage if file doesn't exist
			return &StorageData{
				Version:   storageVersion,
				Profiles:  []ConnectionProfile{},
				Templates: []LogTemplate{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read storage file: %w", err)
	}

	var storage StorageData
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("failed to parse storage file: %w", err)
	}

	return &storage, nil
}

// saveStorage saves the storage data to disk
func saveStorage(storage *StorageData) error {
	path, err := getStoragePath()
	if err != nil {
		return err
	}

	storage.Version = storageVersion
	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize storage: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// generateID generates a unique ID based on timestamp
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ================================================================================
// CONNECTION PROFILES API
// ================================================================================

// GetProfiles returns all saved connection profiles
func (a *App) GetProfiles() ([]ConnectionProfile, error) {
	storage, err := loadStorage()
	if err != nil {
		return nil, err
	}

	// Sort by name
	sort.Slice(storage.Profiles, func(i, j int) bool {
		return storage.Profiles[i].Name < storage.Profiles[j].Name
	})

	return storage.Profiles, nil
}

// SaveProfile saves a new or updates an existing connection profile
func (a *App) SaveProfile(profile ConnectionProfile) (ConnectionProfile, error) {
	storage, err := loadStorage()
	if err != nil {
		return ConnectionProfile{}, err
	}

	now := time.Now().Unix()

	if profile.ID == "" {
		// New profile
		profile.ID = generateID()
		profile.CreatedAt = now
		profile.UpdatedAt = now
		storage.Profiles = append(storage.Profiles, profile)
	} else {
		// Update existing
		found := false
		for i, p := range storage.Profiles {
			if p.ID == profile.ID {
				profile.CreatedAt = p.CreatedAt
				profile.UpdatedAt = now
				storage.Profiles[i] = profile
				found = true
				break
			}
		}
		if !found {
			return ConnectionProfile{}, fmt.Errorf("profile not found: %s", profile.ID)
		}
	}

	if err := saveStorage(storage); err != nil {
		return ConnectionProfile{}, err
	}

	return profile, nil
}

// DeleteProfile deletes a connection profile by ID
func (a *App) DeleteProfile(id string) error {
	storage, err := loadStorage()
	if err != nil {
		return err
	}

	found := false
	for i, p := range storage.Profiles {
		if p.ID == id {
			storage.Profiles = append(storage.Profiles[:i], storage.Profiles[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", id)
	}

	return saveStorage(storage)
}

// ================================================================================
// LOG TEMPLATES API
// ================================================================================

// GetTemplates returns all saved log templates
func (a *App) GetTemplates() ([]LogTemplate, error) {
	storage, err := loadStorage()
	if err != nil {
		return nil, err
	}

	// Sort by name
	sort.Slice(storage.Templates, func(i, j int) bool {
		return storage.Templates[i].Name < storage.Templates[j].Name
	})

	return storage.Templates, nil
}

// SaveTemplate saves a new or updates an existing log template
func (a *App) SaveTemplate(template LogTemplate) (LogTemplate, error) {
	storage, err := loadStorage()
	if err != nil {
		return LogTemplate{}, err
	}

	now := time.Now().Unix()

	if template.ID == "" {
		// New template
		template.ID = generateID()
		template.CreatedAt = now
		template.UpdatedAt = now
		storage.Templates = append(storage.Templates, template)
	} else {
		// Update existing
		found := false
		for i, t := range storage.Templates {
			if t.ID == template.ID {
				template.CreatedAt = t.CreatedAt
				template.UpdatedAt = now
				storage.Templates[i] = template
				found = true
				break
			}
		}
		if !found {
			return LogTemplate{}, fmt.Errorf("template not found: %s", template.ID)
		}
	}

	if err := saveStorage(storage); err != nil {
		return LogTemplate{}, err
	}

	return template, nil
}

// DeleteTemplate deletes a log template by ID
func (a *App) DeleteTemplate(id string) error {
	storage, err := loadStorage()
	if err != nil {
		return err
	}

	found := false
	for i, t := range storage.Templates {
		if t.ID == id {
			storage.Templates = append(storage.Templates[:i], storage.Templates[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("template not found: %s", id)
	}

	return saveStorage(storage)
}

// ================================================================================
// IMPORT/EXPORT API
// ================================================================================

// ExportConfig exports all profiles and templates to a JSON string
func (a *App) ExportConfig() (string, error) {
	storage, err := loadStorage()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to export config: %w", err)
	}

	return string(data), nil
}

// ImportConfig imports profiles and templates from a JSON string
// merge=true will add to existing, merge=false will replace all
func (a *App) ImportConfig(jsonData string, merge bool) error {
	var importedStorage StorageData
	if err := json.Unmarshal([]byte(jsonData), &importedStorage); err != nil {
		return fmt.Errorf("invalid import data: %w", err)
	}

	if merge {
		currentStorage, err := loadStorage()
		if err != nil {
			return err
		}

		// Merge profiles (avoid duplicates by name)
		existingNames := make(map[string]bool)
		for _, p := range currentStorage.Profiles {
			existingNames[p.Name] = true
		}
		for _, p := range importedStorage.Profiles {
			if !existingNames[p.Name] {
				p.ID = generateID() // Generate new ID
				currentStorage.Profiles = append(currentStorage.Profiles, p)
			}
		}

		// Merge templates
		existingTemplateNames := make(map[string]bool)
		for _, t := range currentStorage.Templates {
			existingTemplateNames[t.Name] = true
		}
		for _, t := range importedStorage.Templates {
			if !existingTemplateNames[t.Name] {
				t.ID = generateID()
				currentStorage.Templates = append(currentStorage.Templates, t)
			}
		}

		return saveStorage(currentStorage)
	}

	// Replace all
	return saveStorage(&importedStorage)
}
