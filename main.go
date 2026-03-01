package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/username/caddy-docker-autoproxy/caddy"
	"github.com/username/caddy-docker-autoproxy/config"
	"github.com/username/caddy-docker-autoproxy/docker"
	"github.com/username/caddy-docker-autoproxy/labels"
)

const (
	defaultPollInterval = 30 * time.Second
)

type Poller struct {
	dockerClient    *docker.DockerClient
	caddyClient    *caddy.Client
	pollInterval   time.Duration
	seenContainers map[string]bool
	mu             sync.Mutex
	logger         *slog.Logger
}

func NewPoller(dockerClient *docker.DockerClient, caddyClient *caddy.Client, pollInterval time.Duration) *Poller {
	if pollInterval == 0 {
		pollInterval = defaultPollInterval
	}

	return &Poller{
		dockerClient:    dockerClient,
		caddyClient:     caddyClient,
		pollInterval:    pollInterval,
		seenContainers: make(map[string]bool),
		logger:         slog.Default(),
	}
}

func (p *Poller) Run(ctx context.Context) error {
	p.logger.Info("Starting caddy-docker-autoproxy poller", "interval", p.pollInterval)

	// Do initial poll immediately
	if err := p.poll(ctx); err != nil {
		p.logger.Error("Initial poll failed", "error", err)
	}

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Shutting down poller")
			return ctx.Err()
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				p.logger.Error("Poll failed", "error", err)
			}
		}
	}
}

func (p *Poller) poll(ctx context.Context) error {
	p.logger.Debug("Polling Docker for containers")

	containers, err := p.dockerClient.ListContainersWithLabels(ctx)
	if err != nil {
		return err
	}

	// Filter containers with caddy.enable=true
	var enabledConfigs []labels.CaddyConfig
	currentContainers := make(map[string]bool)

	for _, container := range containers {
		currentContainers[container.ID] = true

		// Parse labels - fsouza client uses Labels map[string]string
		cfg := labels.ParseContainerLabelsSafe(container.Labels)
		if cfg == nil {
			continue
		}

		enabledConfigs = append(enabledConfigs, *cfg)
		// Get container name (remove leading /)
		name := container.Names[0]
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		p.logger.Info("Found enabled container",
			"container", name,
			"host", cfg.Host,
			"port", cfg.Port,
		)
	}

	// Check if containers changed
	changed := false
	if len(enabledConfigs) != len(p.seenContainers) {
		changed = true
	} else {
		for id := range p.seenContainers {
			if !currentContainers[id] {
				changed = true
				break
			}
		}
	}

	// Update seen containers
	p.mu.Lock()
	p.seenContainers = currentContainers
	p.mu.Unlock()

	if !changed {
		p.logger.Debug("No container changes detected")
		return nil
	}

	p.logger.Info("Container configuration changed, updating Caddy")

	// Build Caddy config
	caddyConfig, err := config.BuildReverseProxyConfig(enabledConfigs)
	if err != nil {
		return err
	}

	// Push to Caddy
	if err := p.caddyClient.PushConfig(ctx, caddyConfig); err != nil {
		p.logger.Error("Failed to push config to Caddy", "error", err)
		return err
	}

	p.logger.Info("Successfully updated Caddy configuration")
	return nil
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// CLI flags
	pollInterval := flag.Duration("poll-interval", defaultPollInterval, "Polling interval for checking containers (env: POLL_INTERVAL)")
	caddyURL := flag.String("caddy-url", "http://localhost:2019", "Caddy Admin API URL (env: CADDY_URL)")
	dockerSocket := flag.String("docker-socket", "/var/run/docker.sock", "Docker socket path (env: DOCKER_SOCKET)")
	showHelp := flag.Bool("help", false, "Show help")

	flag.Parse()

	// Override with env vars if set
	if envInterval := os.Getenv("POLL_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			pollInterval = &d
		}
	}
	if envCaddyURL := os.Getenv("CADDY_URL"); envCaddyURL != "" {
		caddyURL = &envCaddyURL
	}
	if envDockerSocket := os.Getenv("DOCKER_SOCKET"); envDockerSocket != "" {
		dockerSocket = &envDockerSocket
	}

	if *showHelp {
		fmt.Fprintf(os.Stderr, "Usage: caddy-docker-autoproxy [options]\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	logger.Info("Starting caddy-docker-autoproxy",
		"poll_interval", *pollInterval,
		"caddy_url", *caddyURL,
		"docker_socket", *dockerSocket)

	// Create Docker client
	dockerClient, err := docker.NewDockerClientWithSocket(context.Background(), *dockerSocket)
	if err != nil {
		logger.Error("Failed to create Docker client", "error", err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	// Create Caddy client with custom URL
	caddyClient := caddy.NewClientWithURL(*caddyURL)

	// Create poller
	poller := NewPoller(dockerClient, caddyClient, *pollInterval)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Run poller
	if err := poller.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("Poller error", "error", err)
		os.Exit(1)
	}

	logger.Info("caddy-docker-autoproxy stopped")
}
