package settings

import "testing"

func TestLoadMissingPreferencesReturnsZeroValue(t *testing.T) {
	configDir := t.TempDir()
	restore := stubUserConfigDir(configDir)
	defer restore()

	preferences, err := Load()
	if err != nil {
		t.Fatalf("load preferences: %v", err)
	}
	if preferences.AskedADHDGate {
		t.Fatal("asked gate = true, want false")
	}
	if preferences.PrimaryForADHD {
		t.Fatal("primary for ADHD = true, want false")
	}
}

func TestSaveAndLoadPreferences(t *testing.T) {
	configDir := t.TempDir()
	restore := stubUserConfigDir(configDir)
	defer restore()

	want := Preferences{
		AskedADHDGate:  true,
		PrimaryForADHD: true,
	}

	if err := Save(want); err != nil {
		t.Fatalf("save preferences: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("load preferences: %v", err)
	}

	if got != want {
		t.Fatalf("preferences = %+v, want %+v", got, want)
	}
}

func stubUserConfigDir(path string) func() {
	original := userConfigDir
	userConfigDir = func() (string, error) {
		return path, nil
	}

	return func() {
		userConfigDir = original
	}
}
