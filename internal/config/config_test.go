package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: esp32c3
monitor:
  baud: 115200
  port: /dev/ttyUSB0
  timestamp: true
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Monitor.Baud != 115200 {
		t.Errorf("expected baud 115200, got %d", cfg.Monitor.Baud)
	}
	if !cfg.Monitor.Timestamp {
		t.Error("expected timestamp true")
	}
}

func TestLoadNotExist(t *testing.T) {
	_, err := Load("/nonexistent/zyrthi.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
