package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jbutlerdev/dev-team/internal/state"
)

func TestLoadConfig(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Error getting user home dir: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "dev-team")
	configPath := filepath.Join(configDir, "config.json")

	// Clean up config file after test
	defer os.Remove(configPath)

	// Test when config file doesn't exist
	appState, err := LoadConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if appState == nil {
		t.Error("AppState should not be nil")
	}

	// Test SaveConfig
	err = SaveConfig()
	if err != nil {
		t.Fatalf("Error saving config: %v", err)
	}

	// Test when config file exists
	appState, err = LoadConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if appState == nil {
		t.Error("AppState should not be nil")
	}
}

func TestSaveConfig(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Error getting user home dir: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "dev-team")
	configPath := filepath.Join(configDir, "config.json")

	// Clean up config file after test
	defer os.Remove(configPath)

	// Create a dummy state
	state.State = &state.AppState{}

	// Save the config
	err = SaveConfig()
	if err != nil {
		t.Fatalf("Error saving config: %v", err)
	}

	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}
}
