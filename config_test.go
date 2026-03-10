package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
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

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Monitor.Baud != 115200 {
		t.Errorf("expected baud 115200, got %d", cfg.Monitor.Baud)
	}
	if cfg.Monitor.Port != "/dev/ttyUSB0" {
		t.Errorf("expected port '/dev/ttyUSB0', got %s", cfg.Monitor.Port)
	}
	if !cfg.Monitor.Timestamp {
		t.Error("expected timestamp true")
	}
}

func TestLoadConfigNotExist(t *testing.T) {
	_, err := loadConfig("/nonexistent/zyrthi.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: [invalid yaml
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zyrthi.yaml")

	content := `platform: esp32
chip: esp32c3
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}

	// 检查默认值
	if cfg.Monitor.Baud != 0 {
		t.Errorf("expected default baud 0, got %d", cfg.Monitor.Baud)
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Platform: "esp32",
		Chip:     "esp32c3",
		Monitor: MonitorConfig{
			Baud:      921600,
			Port:      "/dev/ttyACM0",
			Timestamp: true,
		},
	}

	if cfg.Platform != "esp32" {
		t.Errorf("expected platform 'esp32', got %s", cfg.Platform)
	}
	if cfg.Monitor.Baud != 921600 {
		t.Errorf("expected baud 921600, got %d", cfg.Monitor.Baud)
	}
	if cfg.Monitor.Port != "/dev/ttyACM0" {
		t.Errorf("expected port '/dev/ttyACM0', got %s", cfg.Monitor.Port)
	}
	if !cfg.Monitor.Timestamp {
		t.Error("expected timestamp true")
	}
}

func TestMonitorConfigStruct(t *testing.T) {
	mc := MonitorConfig{
		Baud:      115200,
		Port:      "/dev/ttyUSB0",
		Timestamp: true,
	}

	if mc.Baud != 115200 {
		t.Errorf("expected baud 115200, got %d", mc.Baud)
	}
	if mc.Port != "/dev/ttyUSB0" {
		t.Errorf("expected port '/dev/ttyUSB0', got %s", mc.Port)
	}
	if !mc.Timestamp {
		t.Error("expected timestamp true")
	}
}

func TestOptionsStruct(t *testing.T) {
	opts := Options{
		Timestamp: true,
		Hex:       false,
		LogFile:   "/tmp/monitor.log",
		Filter:    "error",
	}

	if !opts.Timestamp {
		t.Error("expected timestamp true")
	}
	if opts.Hex {
		t.Error("expected hex false")
	}
	if opts.LogFile != "/tmp/monitor.log" {
		t.Errorf("expected log file '/tmp/monitor.log', got %s", opts.LogFile)
	}
	if opts.Filter != "error" {
		t.Errorf("expected filter 'error', got %s", opts.Filter)
	}
}
