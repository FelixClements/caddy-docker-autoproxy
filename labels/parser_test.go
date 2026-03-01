package labels

import (
	"testing"
)

func TestParseContainerLabels_Enabled(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "true",
		"caddy.host":  "example.com",
		"caddy.port":  "8080",
	}

	config, err := ParseContainerLabels(labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	if !config.Enable {
		t.Error("expected Enable to be true")
	}

	if config.Host != "example.com" {
		t.Errorf("expected host 'example.com', got '%s'", config.Host)
	}

	if config.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Port)
	}
}

func TestParseContainerLabels_Disabled(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "false",
		"caddy.host":  "example.com",
		"caddy.port":  "8080",
	}

	config, err := ParseContainerLabels(labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config != nil {
		t.Error("expected nil config for disabled container")
	}
}

func TestParseContainerLabels_NoLabels(t *testing.T) {
	labels := map[string]string{}

	config, err := ParseContainerLabels(labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config != nil {
		t.Error("expected nil config for container without caddy labels")
	}
}

func TestParseContainerLabels_MissingHost(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "true",
		"caddy.port":   "8080",
	}

	config, err := ParseContainerLabels(labels)
	if err == nil {
		t.Fatal("expected error for missing host")
	}

	if config != nil {
		t.Error("expected nil config")
	}
}

func TestParseContainerLabels_MissingPort(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "true",
		"caddy.host":   "example.com",
	}

	config, err := ParseContainerLabels(labels)
	if err == nil {
		t.Fatal("expected error for missing port")
	}

	if config != nil {
		t.Error("expected nil config")
	}
}

func TestParseContainerLabels_InvalidPort(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "true",
		"caddy.host":   "example.com",
		"caddy.port":   "not-a-number",
	}

	config, err := ParseContainerLabels(labels)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}

	if config != nil {
		t.Error("expected nil config")
	}
}

func TestParseContainerLabels_WithPath(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "true",
		"caddy.host":   "example.com",
		"caddy.port":   "8080",
		"caddy.path":   "/api",
	}

	config, err := ParseContainerLabels(labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	if config.Path != "/api" {
		t.Errorf("expected path '/api', got '%s'", config.Path)
	}
}

func TestParseContainerLabelsSafe_Enabled(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "true",
		"caddy.host":   "example.com",
		"caddy.port":   "8080",
	}

	config := ParseContainerLabelsSafe(labels)
	if config == nil {
		t.Error("expected config, got nil")
	}
}

func TestParseContainerLabelsSafe_Disabled(t *testing.T) {
	labels := map[string]string{
		"caddy.enable": "false",
		"caddy.host":   "example.com",
		"caddy.port":   "8080",
	}

	config := ParseContainerLabelsSafe(labels)
	if config != nil {
		t.Error("expected nil config for disabled")
	}
}

func TestParseContainerLabels_WithAddress(t *testing.T) {
	labels := map[string]string{
		"caddy.enable":  "true",
		"caddy.host":    "backend.example.com",
		"caddy.port":    "8080",
		"caddy.address": "example.com",
	}

	config, err := ParseContainerLabels(labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	if config.Address != "example.com" {
		t.Errorf("expected address 'example.com', got '%s'", config.Address)
	}
}

func TestParseContainerLabels_AddressEmpty(t *testing.T) {
	labels := map[string]string{
		"caddy.enable":  "true",
		"caddy.host":    "example.com",
		"caddy.port":    "8080",
		"caddy.address": "",
	}

	config, err := ParseContainerLabels(labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	// Empty address should be treated as missing
	if config.Address != "" {
		t.Errorf("expected empty address, got '%s'", config.Address)
	}
}
