package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func applicationStatePath() string {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "state.json"
	}

	return filepath.Join(configDirectory, "ds1-death-counter", "state.json")
}

func loadApplicationState(applicationState *applicationState) {
	data, err := os.ReadFile(applicationStatePath())
	if err != nil {
		return
	}

	var savedState stateSnapshot
	if err := json.Unmarshal(data, &savedState); err != nil {
		return
	}

	applicationState.mu.Lock()
	defer applicationState.mu.Unlock()

	if savedState.Label != "" {
		applicationState.Label = savedState.Label
	}
	if savedState.Killed >= 0 && savedState.Killed <= totalBossCount {
		applicationState.Killed = savedState.Killed
	}
}

func saveApplicationState(applicationState *applicationState) error {
	data, err := json.Marshal(applicationState.snapshot())
	if err != nil {
		return err
	}

	path := applicationStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
