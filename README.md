# caddy-docker-autoproxy

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

Auto-configure Caddy reverse proxy based on Docker container labels.

## Features

- **Automatic Discovery**: Polls Docker containers and detects new ones with Caddy labels
- **Label-Based Configuration**: Simple Docker labels to enable reverse proxy
- **Hot Reloading**: Automatically updates Caddy config when containers change
- **Graceful Shutdown**: Handles SIGINT/SIGTERM properly

## Installation

### From Binary

Download the latest release from [GitHub Releases](https://github.com/username/caddy-docker-autoproxy/releases):

```bash
curl -L -o caddy-docker-autoproxy https://github.com/username/caddy-docker-autoproxy/releases/latest/download/caddy-docker-autoproxy
chmod +x caddy-docker-autoproxy
```

### From Source

```bash
git clone https://github.com/username/caddy-docker-autoproxy.git
cd caddy-docker-autoproxy
go build -o caddy-docker-autoproxy .
```

## Usage

### Quick Start

```bash
./caddy-docker-autoproxy
```

### With Docker

```bash
docker run -d \
  --name caddy-autoproxy \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e CADDY_URL=http://caddy:2019 \
  your-repo/caddy-docker-autoproxy
```

## Docker Labels

Add these labels to your Docker containers:

| Label | Required | Description |
|-------|----------|-------------|
| `caddy.enable` | Yes | Set to `true` to enable proxy |
| `caddy.host` | Yes | Backend upstream host (where to proxy to) |
| `caddy.port` | Yes | Backend upstream port number |
| `caddy.path` | No | Path prefix for route (e.g., `/api`) |
| `caddy.address` | Yes | Public-facing hostname (e.g., `example.com`). Caddy auto-HTTPS handles ports 80/443 |

### Example

```yaml
services:
  myapp:
    image: nginx:latest
    labels:
      caddy.enable: "true"
      caddy.host: "backend.example.com"
      caddy.port: "80"
      caddy.address: "192.168.1.10"
      # Optional: caddy.path: "/app"
```

## Configuration

### Command Line Flags

| Flag | Env Variable | Default | Description |
|------|--------------|---------|-------------|
| `--poll-interval` | `POLL_INTERVAL` | `30s` | Polling interval |
| `--caddy-url` | `CADDY_URL` | `http://localhost:2019` | Caddy Admin API URL |
| `--docker-socket` | `DOCKER_SOCKET` | `/var/run/docker.sock` | Docker socket path |

### Example with Custom Settings

```bash
./caddy-docker-autoproxy \
  --poll-interval=10s \
  --caddy-url=http://localhost:2019 \
  --docker-socket=/var/run/docker.sock
```

## How It Works

1. Polls Docker for running containers every 30 seconds
2. Filters containers with `caddy.enable=true` label
3. Reads `caddy.host`, `caddy.port`, optional `caddy.path`, and `caddy.address`
4. Generates Caddy JSON reverse proxy config with host-based matchers
5. Pushes config to Caddy Admin API

## Architecture

```
caddy-docker-autoproxy
├── main.go           # Entry point and polling loop
├── docker/           # Docker API client
├── caddy/           # Caddy Admin API client
├── config/          # JSON config builder
└── labels/          # Label parser
```

## Development

### Run Tests

```bash
go test -v ./...
```

### Build

```bash
go build -o caddy-docker-autoproxy .
```

## License

MIT License - see [LICENSE](LICENSE) for details.
