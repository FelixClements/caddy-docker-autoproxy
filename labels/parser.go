package labels

import (
	"errors"
	"strconv"
)

// CaddyConfig represents the parsed Caddy configuration from container labels
type CaddyConfig struct {
	Enable  bool
	Host    string
	Port    int
	Path    string // optional
	Address string // optional
}

// ParseContainerLabels parses Docker labels and extracts Caddy configuration
func ParseContainerLabels(labels map[string]string) (*CaddyConfig, error) {
	// Check if caddy.enable is set to "true"
	enableVal, exists := labels["caddy.enable"]
	if !exists || enableVal != "true" {
		return nil, nil // Container not enabled for Caddy proxy
	}

	config := &CaddyConfig{
		Enable: true,
	}

	// Parse caddy.host (required)
	host, exists := labels["caddy.host"]
	if !exists || host == "" {
		return nil, errors.New("caddy.host is required when caddy.enable=true")
	}
	config.Host = host

	// Parse caddy.port (required)
	portStr, exists := labels["caddy.port"]
	if !exists || portStr == "" {
		return nil, errors.New("caddy.port is required when caddy.enable=true")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("caddy.port must be a valid integer")
	}
	config.Port = port

	// Parse caddy.path (optional)
	if path, exists := labels["caddy.path"]; exists {
		config.Path = path
	}

	// Parse caddy.address (optional)
	if address, exists := labels["caddy.address"]; exists && address != "" {
		config.Address = address
	}

	return config, nil
}

// ParseContainerLabelsSafe is a wrapper that returns nil for non-enabled containers
// instead of an error
func ParseContainerLabelsSafe(labels map[string]string) *CaddyConfig {
	config, err := ParseContainerLabels(labels)
	if err != nil {
		return nil
	}
	return config
}
