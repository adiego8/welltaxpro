package server

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SslMode  string `yaml:"sslmode"`
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowedOrigins"`
	AllowedMethods   []string `yaml:"allowedMethods"`
	AllowedHeaders   []string `yaml:"allowedHeaders"`
	AllowCredentials bool     `yaml:"allowCredentials"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type FirebaseConfig struct {
	APIKey             string `yaml:"apiKey"`
	ServiceAccountPath string `yaml:"serviceAccountPath"`
}

type SendGridConfig struct {
	APIKey           string `yaml:"apiKey"`
	DefaultFromEmail string `yaml:"defaultFromEmail"`
	DefaultFromName  string `yaml:"defaultFromName"`
}

type PortalConfig struct {
	JWTSecret string `yaml:"jwtSecret"`
	BaseURL   string `yaml:"baseURL"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Cors     CORSConfig     `yaml:"cors"`
	Firebase FirebaseConfig `yaml:"firebase"`
	SendGrid SendGridConfig `yaml:"sendgrid"`
	Portal   PortalConfig   `yaml:"portal"`
}

func getConfiguration(args *Arguments) (*Config, error) {
	configPath := args.ConfigPath
	if configPath == "" {
		configPath = "config.yaml"
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
