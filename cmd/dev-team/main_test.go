package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(t *testing.T) {
	// Get the user's home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Error getting user home directory: %v", err)
	}

	// Construct the config file path.
	configPath := filepath.Join(homeDir, ".config", "dev-team", "config.json")

	// Delete the config file if it exists.
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			t.Fatalf("Error deleting config file: %v", err)
		}
	}

	// Call the main function.
	main()

	// Check if the config file exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created: %v", err)
	}
}
