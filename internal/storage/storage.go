// Package storage provides persistent JSON storage for connection profiles
// and log templates, shared by the GUI and the CLI.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ConnectionProfile represents a saved connection configuration
type ConnectionProfile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Address        string `json:"address"`
	Port           string `json:"port"`
	Protocol       string `json:"protocol"`
	Method         string `json:"framingMethod"`
	UseTLS         bool   `json:"useTls"`
	TLSVerify      bool   `json:"tlsVerify"`
	CACertPath     string `json:"caCertPath,omitempty"`
	ClientCertPath string `json:"clientCertPath,omitempty"`
	ClientKeyPath  string `json:"clientKeyPath,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
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
	// MessageFormat is optional so templates stored before it existed still
	// decode; an empty value falls back to UseRFC5424.
	MessageFormat string `json:"messageFormat,omitempty"`
	UseRFC5424    bool   `json:"useRfc5424"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// Data represents the complete storage file structure
type Data struct {
	Version   int                 `json:"version"`
	Profiles  []ConnectionProfile `json:"profiles"`
	Templates []LogTemplate       `json:"templates"`
}

const (
	// StorageVersion is the schema version of the storage file
	StorageVersion = 1

	// StorageFileName is the name of the JSON config file inside the app folder
	StorageFileName = "sendlog-config.json"

	// AppFolderName is the name of the application's configuration folder
	AppFolderName = "SendLog-Syslog"
)

// Path returns the path to the storage file
// Uses XDG config on Linux, AppData on Windows, Library on macOS
func Path() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	appDir := filepath.Join(configDir, AppFolderName)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(appDir, StorageFileName), nil
}

// Load loads the storage data from disk
func Load() (*Data, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty storage if file doesn't exist
			return &Data{
				Version:   StorageVersion,
				Profiles:  []ConnectionProfile{},
				Templates: []LogTemplate{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read storage file: %w", err)
	}

	var s Data
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse storage file: %w", err)
	}

	return &s, nil
}

// Save writes the storage data to disk
func Save(s *Data) error {
	path, err := Path()
	if err != nil {
		return err
	}

	s.Version = StorageVersion
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize storage: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// GenerateID generates a unique ID based on timestamp
func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetProfiles returns all saved connection profiles
func GetProfiles() ([]ConnectionProfile, error) {
	s, err := Load()
	if err != nil {
		return nil, err
	}

	sort.Slice(s.Profiles, func(i, j int) bool {
		return s.Profiles[i].Name < s.Profiles[j].Name
	})

	return s.Profiles, nil
}

// SaveProfile saves a new or updates an existing connection profile
func SaveProfile(profile ConnectionProfile) (ConnectionProfile, error) {
	s, err := Load()
	if err != nil {
		return ConnectionProfile{}, err
	}

	now := time.Now().Unix()

	if profile.ID == "" {
		profile.ID = GenerateID()
		profile.CreatedAt = now
		profile.UpdatedAt = now
		s.Profiles = append(s.Profiles, profile)
	} else {
		found := false
		for i, p := range s.Profiles {
			if p.ID == profile.ID {
				profile.CreatedAt = p.CreatedAt
				profile.UpdatedAt = now
				s.Profiles[i] = profile
				found = true
				break
			}
		}
		if !found {
			return ConnectionProfile{}, fmt.Errorf("profile not found: %s", profile.ID)
		}
	}

	if err := Save(s); err != nil {
		return ConnectionProfile{}, err
	}

	return profile, nil
}

// DeleteProfile deletes a connection profile by ID
func DeleteProfile(id string) error {
	s, err := Load()
	if err != nil {
		return err
	}

	found := false
	for i, p := range s.Profiles {
		if p.ID == id {
			s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", id)
	}

	return Save(s)
}

// GetTemplates returns all saved log templates
func GetTemplates() ([]LogTemplate, error) {
	s, err := Load()
	if err != nil {
		return nil, err
	}

	sort.Slice(s.Templates, func(i, j int) bool {
		return s.Templates[i].Name < s.Templates[j].Name
	})

	return s.Templates, nil
}

// SaveTemplate saves a new or updates an existing log template
func SaveTemplate(template LogTemplate) (LogTemplate, error) {
	s, err := Load()
	if err != nil {
		return LogTemplate{}, err
	}

	now := time.Now().Unix()

	if template.ID == "" {
		template.ID = GenerateID()
		template.CreatedAt = now
		template.UpdatedAt = now
		s.Templates = append(s.Templates, template)
	} else {
		found := false
		for i, t := range s.Templates {
			if t.ID == template.ID {
				template.CreatedAt = t.CreatedAt
				template.UpdatedAt = now
				s.Templates[i] = template
				found = true
				break
			}
		}
		if !found {
			return LogTemplate{}, fmt.Errorf("template not found: %s", template.ID)
		}
	}

	if err := Save(s); err != nil {
		return LogTemplate{}, err
	}

	return template, nil
}

// DeleteTemplate deletes a log template by ID
func DeleteTemplate(id string) error {
	s, err := Load()
	if err != nil {
		return err
	}

	found := false
	for i, t := range s.Templates {
		if t.ID == id {
			s.Templates = append(s.Templates[:i], s.Templates[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("template not found: %s", id)
	}

	return Save(s)
}

// Export returns all profiles and templates as a JSON string
func Export() (string, error) {
	s, err := Load()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to export config: %w", err)
	}

	return string(data), nil
}

// Import loads profiles and templates from a JSON string.
// merge=true adds to existing; merge=false replaces all.
func Import(jsonData string, merge bool) error {
	var imported Data
	if err := json.Unmarshal([]byte(jsonData), &imported); err != nil {
		return fmt.Errorf("invalid import data: %w", err)
	}

	if merge {
		current, err := Load()
		if err != nil {
			return err
		}

		existingNames := make(map[string]bool)
		for _, p := range current.Profiles {
			existingNames[p.Name] = true
		}
		for _, p := range imported.Profiles {
			if !existingNames[p.Name] {
				p.ID = GenerateID()
				current.Profiles = append(current.Profiles, p)
			}
		}

		existingTemplateNames := make(map[string]bool)
		for _, t := range current.Templates {
			existingTemplateNames[t.Name] = true
		}
		for _, t := range imported.Templates {
			if !existingTemplateNames[t.Name] {
				t.ID = GenerateID()
				current.Templates = append(current.Templates, t)
			}
		}

		return Save(current)
	}

	return Save(&imported)
}
