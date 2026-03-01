package config

import (
	"encoding/json"
	"fmt"

	"github.com/username/caddy-docker-autoproxy/labels"
)

// ContainerConfig represents a container to be proxied
type ContainerConfig struct {
	Host    string
	Port    int
	Path    string
	Address string
}

// BuildReverseProxyConfig builds Caddy JSON config for reverse proxy
func BuildReverseProxyConfig(containers []labels.CaddyConfig) (map[string]interface{}, error) {
	if len(containers) == 0 {
		// Return empty config when no containers
		return buildEmptyConfig(), nil
	}

	routes := make([]interface{}, 0, len(containers))

	for i, c := range containers {
		route := buildRoute(c, i)
		routes = append(routes, route)
	}

	config := map[string]interface{}{
		"apps": map[string]interface{}{
			"http": map[string]interface{}{
				"servers": map[string]interface{}{
					"auto_proxy": map[string]interface{}{
						"routes": routes,
					},
				},
			},
		},
	}

	return config, nil
}

func buildEmptyConfig() map[string]interface{} {
	return map[string]interface{}{
		"apps": map[string]interface{}{
			"http": map[string]interface{}{
				"servers": map[string]interface{}{
					"auto_proxy": map[string]interface{}{
						"routes": []interface{}{},
					},
				},
			},
		},
	}
}

func buildRoute(c labels.CaddyConfig, index int) map[string]interface{} {
	handle := map[string]interface{}{
		"handler":     "reverse_proxy",
		"upstreams":   []map[string]string{{"dial": fmt.Sprintf("%s:%d", c.Host, c.Port)}},
	}

	// Build matchers based on Address and Path
	if c.Address != "" || c.Path != "" {
		matcher := map[string]interface{}{}

		if c.Address != "" {
			matcher["host"] = []string{c.Address}
		}
		if c.Path != "" {
			matcher["path"] = []string{c.Path + "/*"}
		}

		return map[string]interface{}{
			"match":  []map[string]interface{}{matcher},
			"handle": []interface{}{handle},
		}
	}

	// Default: catch-all route
	return map[string]interface{}{
		"handle": []interface{}{handle},
	}
}

// ToJSON converts the config to JSON bytes
func ToJSON(config map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}

// ToJSONString converts the config to a JSON string
func ToJSONString(config map[string]interface{}) (string, error) {
	data, err := ToJSON(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
