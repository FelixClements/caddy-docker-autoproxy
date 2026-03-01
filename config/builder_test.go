package config

import (
	"encoding/json"
	"testing"

	"github.com/username/caddy-docker-autoproxy/labels"
)

func TestBuildReverseProxyConfig_SingleContainer(t *testing.T) {
	containers := []labels.CaddyConfig{
		{
			Enable: true,
			Host:   "example.com",
			Port:   8080,
		},
	}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	// Verify JSON structure is valid
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestBuildReverseProxyConfig_MultipleContainers(t *testing.T) {
	containers := []labels.CaddyConfig{
		{
			Enable: true,
			Host:   "example.com",
			Port:   8080,
		},
		{
			Enable: true,
			Host:   "api.example.com",
			Port:   3000,
		},
	}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := json.Marshal(config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	apps := result["apps"].(map[string]interface{})
	httpApps := apps["http"].(map[string]interface{})
	servers := httpApps["servers"].(map[string]interface{})
	autoProxy := servers["auto_proxy"].(map[string]interface{})
	routes := autoProxy["routes"].([]interface{})

	if len(routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routes))
	}
}

func TestBuildReverseProxyConfig_Empty(t *testing.T) {
	containers := []labels.CaddyConfig{}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	// Verify JSON structure is valid
	data, _ := json.Marshal(config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Verify basic structure exists
	apps := result["apps"].(map[string]interface{})
	_ = apps["http"]
}

func TestBuildReverseProxyConfig_WithPath(t *testing.T) {
	containers := []labels.CaddyConfig{
		{
			Enable: true,
			Host:   "example.com",
			Port:   8080,
			Path:   "/api",
		},
	}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := json.Marshal(config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Check that path matcher is included
	apps := result["apps"].(map[string]interface{})
	httpApps := apps["http"].(map[string]interface{})
	servers := httpApps["servers"].(map[string]interface{})
	autoProxy := servers["auto_proxy"].(map[string]interface{})
	routes := autoProxy["routes"].([]interface{})
	route := routes[0].(map[string]interface{})

	_, ok := route["match"]
	if !ok {
		t.Error("expected 'match' in route for path-based routing")
	}
}

func TestToJSON(t *testing.T) {
	config := map[string]interface{}{
		"test": "value",
	}

	data, err := ToJSON(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestToJSONString(t *testing.T) {
	config := map[string]interface{}{
		"test": "value",
	}

	str, err := ToJSONString(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(str) == 0 {
		t.Error("expected non-empty string")
	}
}

func TestBuildReverseProxyConfig_WithAddress(t *testing.T) {
	containers := []labels.CaddyConfig{
		{
			Enable:  true,
			Host:    "backend.example.com",
			Port:    8080,
			Address: "example.com",
		},
	}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := json.Marshal(config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Check that host matcher is included
	apps := result["apps"].(map[string]interface{})
	httpApps := apps["http"].(map[string]interface{})
	servers := httpApps["servers"].(map[string]interface{})
	autoProxy := servers["auto_proxy"].(map[string]interface{})
	routes := autoProxy["routes"].([]interface{})
	route := routes[0].(map[string]interface{})

	match, ok := route["match"]
	if !ok {
		t.Error("expected 'match' in route for address-based routing")
	}

	matches := match.([]interface{})
	if len(matches) == 0 {
		t.Error("expected non-empty match array")
	}

	matcher := matches[0].(map[string]interface{})
	hosts, ok := matcher["host"]
	if !ok {
		t.Error("expected 'host' matcher")
	}
	hostList := hosts.([]string)
	if len(hostList) == 0 || hostList[0] != "example.com" {
		t.Errorf("expected host 'example.com', got %v", hostList)
	}
}

func TestBuildReverseProxyConfig_AddressEmpty_NoHostMatcher(t *testing.T) {
	containers := []labels.CaddyConfig{
		{
			Enable: true,
			Host:   "example.com",
			Port:   8080,
			// Address not set
		},
	}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := json.Marshal(config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Check that no host matcher is included (backward compatible)
	apps := result["apps"].(map[string]interface{})
	httpApps := apps["http"].(map[string]interface{})
	servers := httpApps["servers"].(map[string]interface{})
	autoProxy := servers["auto_proxy"].(map[string]interface{})
	routes := autoProxy["routes"].([]interface{})
	route := routes[0].(map[string]interface{})

	_, ok := route["match"]
	if ok {
		t.Error("expected no 'match' in route when address is empty (backward compatible)")
	}
}

func TestBuildReverseProxyConfig_WithAddressAndPath(t *testing.T) {
	containers := []labels.CaddyConfig{
		{
			Enable:  true,
			Host:    "backend.example.com",
			Port:    8080,
			Address: "example.com",
			Path:    "/api",
		},
	}

	config, err := BuildReverseProxyConfig(containers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := json.Marshal(config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Check that both host and path matchers are included
	apps := result["apps"].(map[string]interface{})
	httpApps := apps["http"].(map[string]interface{})
	servers := httpApps["servers"].(map[string]interface{})
	autoProxy := servers["auto_proxy"].(map[string]interface{})
	routes := autoProxy["routes"].([]interface{})
	route := routes[0].(map[string]interface{})

	match, ok := route["match"]
	if !ok {
		t.Error("expected 'match' in route")
	}

	matches := match.([]interface{})
	matcher := matches[0].(map[string]interface{})

	// Check host matcher
	hosts, ok := matcher["host"]
	if !ok {
		t.Error("expected 'host' matcher")
	}
	hostList := hosts.([]string)
	if hostList[0] != "example.com" {
		t.Errorf("expected host 'example.com', got %v", hostList)
	}

	// Check path matcher
	paths, ok := matcher["path"]
	if !ok {
		t.Error("expected 'path' matcher")
	}
	pathList := paths.([]string)
	if pathList[0] != "/api/*" {
		t.Errorf("expected path '/api/*', got %v", pathList)
	}
}
