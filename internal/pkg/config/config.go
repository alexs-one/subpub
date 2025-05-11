package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"grpc_port"`
}

type Logger struct {
	Level string `yaml:"level"`
}

type Config struct {
	Server Server `yaml:"service"`
	Logger Logger `yaml:"logger"`
}

func LoadConfig(filename string) (*Config, error) {
	fileClean := filepath.Clean(filename)
	f, err := os.Open(fileClean)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	config := &Config{}
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
