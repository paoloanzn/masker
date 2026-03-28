package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	appDirName      = "masker"
	preferencesFile = "preferences.json"
)

var userConfigDir = os.UserConfigDir

type Preferences struct {
	AskedADHDGate  bool `json:"asked_adhd_gate"`
	PrimaryForADHD bool `json:"primary_for_adhd"`
}

func Load() (Preferences, error) {
	path, err := preferencesPath()
	if err != nil {
		return Preferences{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Preferences{}, nil
		}
		return Preferences{}, err
	}

	var preferences Preferences
	if err := json.Unmarshal(data, &preferences); err != nil {
		return Preferences{}, err
	}

	return preferences, nil
}

func Save(preferences Preferences) error {
	path, err := preferencesPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func preferencesPath() (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, appDirName, preferencesFile), nil
}
