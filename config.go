package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Settings represents the configuration for the application
type Settings struct {
	JiraBaseURL      string `yaml:"jira_base_url"`
	JiraEmailOrUser  string `yaml:"jira_email"`
	JiraAPIToken     string `yaml:"jira_api_token"`
	Timezone         string `yaml:"timezone"`
	APIVersion       string `yaml:"api_version"`
	LogLevel         string `yaml:"log_level"`
	DefaultsConfig   `yaml:"defaults"`
}

// DefaultsConfig represents the defaults section of the config
type DefaultsConfig struct {
	CategoryAliases map[string]string `yaml:"category_aliases"`
}

// LoadSettings loads the application settings from the config file
func LoadSettings() (*Settings, error) {
	configPath := findConfigFile()
	if configPath == "" {
		return nil, fmt.Errorf("no config file found")
	}

	fmt.Printf("[info] Using config file: %s\n", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	settings := &Settings{}
	err = yaml.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if they exist
	if baseURL := os.Getenv("JIRA_BASE_URL"); baseURL != "" {
		settings.JiraBaseURL = baseURL
	}
	if email := os.Getenv("JIRA_EMAIL"); email != "" {
		settings.JiraEmailOrUser = email
	}
	if token := os.Getenv("JIRA_API_TOKEN"); token != "" {
		settings.JiraAPIToken = token
	}
	if timezone := os.Getenv("TIMEZONE"); timezone != "" {
		settings.Timezone = timezone
	}
	if apiVersion := os.Getenv("JIRA_API_VERSION"); apiVersion != "" {
		settings.APIVersion = apiVersion
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		settings.LogLevel = logLevel
	}

	// Validate required settings
	if settings.JiraBaseURL == "" {
		return nil, fmt.Errorf("missing JIRA_BASE_URL")
	}
	if settings.JiraEmailOrUser == "" {
		return nil, fmt.Errorf("missing JIRA_EMAIL")
	}
	if settings.JiraAPIToken == "" {
		return nil, fmt.Errorf("missing JIRA_API_TOKEN")
	}
	if settings.Timezone == "" {
		settings.Timezone = "Europe/London"
	}
	if settings.APIVersion == "" {
		settings.APIVersion = "3"
	}
	if settings.LogLevel == "" {
		settings.LogLevel = "info"
	}

	return settings, nil
}

// findConfigFile searches for the config file in various locations
func findConfigFile() string {
	// Check environment variable
	if envPath := os.Getenv("WORKLOG_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// Check current directory
	if pwd, err := os.Getwd(); err == nil {
		cwdConfig := filepath.Join(pwd, "worklog_config.yaml")
		if _, err := os.Stat(cwdConfig); err == nil {
			return cwdConfig
		}
	}

	// Check executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		execConfig := filepath.Join(execDir, "worklog_config.yaml")
		if _, err := os.Stat(execConfig); err == nil {
			return execConfig
		}
	}

	// Check user's home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		homeConfig := filepath.Join(homeDir, ".worklog_config.yaml")
		if _, err := os.Stat(homeConfig); err == nil {
			return homeConfig
		}
	}

	// Check system-wide directory
	systemConfig := "/etc/jira-worklogger/worklog_config.yaml"
	if _, err := os.Stat(systemConfig); err == nil {
		return systemConfig
	}

	fmt.Println("[warn] No config file found. Checking: environment variable WORKLOG_CONFIG, ./worklog_config.yaml, executable_dir/worklog_config.yaml, ~/.worklog_config.yaml, /etc/jira-worklogger/worklog_config.yaml")
	return ""
}
