package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config zyrthi.yaml 配置结构
type Config struct {
	Platform string        `yaml:"platform"`
	Chip     string        `yaml:"chip"`
	Monitor  MonitorConfig `yaml:"monitor"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Baud      int    `yaml:"baud"`
	Port      string `yaml:"port"`
	Timestamp bool   `yaml:"timestamp"`
}

// loadConfig 从 zyrthi.yaml 加载配置
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
