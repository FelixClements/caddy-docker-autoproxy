package docker

import (
	"context"
	"errors"
	"os"

	dockerclient "github.com/fsouza/go-dockerclient"
)

// DockerClient wraps the Docker client
type DockerClient struct {
	cli *dockerclient.Client
}

// NewDockerClient creates a new Docker client using DOCKER_SOCKET env var or default
func NewDockerClient(ctx context.Context) (*DockerClient, error) {
	socketPath := os.Getenv("DOCKER_SOCKET")
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}
	return NewDockerClientWithSocket(ctx, socketPath)
}

// NewDockerClientWithSocket creates a new Docker client with a specific socket path
func NewDockerClientWithSocket(ctx context.Context, socketPath string) (*DockerClient, error) {
	if _, err := os.Stat(socketPath); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("docker socket not found: " + socketPath)
	}

	cli, err := dockerclient.NewClient("unix://" + socketPath)
	if err != nil {
		return nil, err
	}

	return &DockerClient{cli: cli}, nil
}

// Close closes the Docker client
func (d *DockerClient) Close() error {
	return nil
}

// ListContainersWithLabels returns running containers with their labels
func (d *DockerClient) ListContainersWithLabels(ctx context.Context) ([]dockerclient.APIContainers, error) {
	containers, err := d.cli.ListContainers(dockerclient.ListContainersOptions{
		All: false,
	})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// ContainerInfo returns detailed info about a container
func (d *DockerClient) ContainerInfo(ctx context.Context, containerID string) (*dockerclient.Container, error) {
	return d.cli.InspectContainer(containerID)
}
