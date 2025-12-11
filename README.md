# mockgrid

[![Security Scan](https://github.com/musturu/mockgrid/actions/workflows/scan.yml/badge.svg?branch=main)](https://github.com/musturu/mockgrid/actions/workflows/scan.yml)
[![License](https://img.shields.io/github/license/musturu/mockgrid)](LICENSE)

mockgrid is a small, mock email server that mimics parts of SendGrid's API so you can test sending emails and template rendering locally without talking to the real SendGrid service.

## Scope
- Provide an HTTP API compatible with common SendGrid flows used by applications (mail send, templates, tracking).
- Allow sending emails through a real SMTP server configured by the operator (so you can wire tests to a local or remote SMTP endpoint).
- Support template rendering from local templates or (optionally) a SendGrid-like template provider.
- Provide attachment handling for testing file uploads and delivery.

## Key functionalities
- HTTP endpoint that accepts SendGrid-like POST requests and forwards messages via SMTP.
- Template rendering engine with support for local template directories and remote templates.
- Attachment handling with secure, temporary storage.
- Tracking pixel support for open tracking (test-friendly).
- Configurable via environment variables, config file, or CLI flags.

## Development status
This project is actively under development. Features, config options, and APIs are subject to change. Use it for testing and experimentation, and expect behavior to evolve.

## Credits
Special thanks to `ykanezawa` and their `sendgrid-dev` (https://github.com/yKanazawa/sendgrid-dev)  repository — their work was a helpful starting point for this project.

## Building and Running

### Prerequisites
- Go 1.25 or later
- Docker (optional, for containerized deployment)

### Build the binary
```bash
make build
```
This compiles the mockgrid binary to `./mockgrid`.

### Run tests
```bash
make test
```
Runs the complete test suite including unit tests, contract tests, and specific store tests.

### Clean build artifacts
```bash
make clean
```
Removes the compiled binary and cleans build cache.

### Run locally
```bash
make run
```
Starts the mockgrid server. Configure it with environment variables or a config file (see `config.example.yaml`).

### Build and run Docker image
```bash
# Build the Docker image
docker build -f deploy/Dockerfile -t mockgrid:latest .

# Run the container
docker run -p 8080:8080 \
  -e SMTP_SERVER=smtp.example.com \
  -e SMTP_PORT=25 \
  -e SENDGRID_KEY=your-api-key \
  mockgrid:latest
```

The Dockerfile in `deploy/` builds a minimal scratch image with only the compiled binary and CA certificates.

### Docker database initialization

The image runs `/docker-entrypoint-initdb.d/*` before it starts the HTTP server, making it easy to seed storage.

- Drop shell scripts (`*.sh`) into the directory to run arbitrary commands inside the container before the service starts.
- Drop SQL files (`*.sql`) to execute them against the configured SQLite database (requires `STORAGE_TYPE=sqlite`).
- Drop JSON files (`*.json`) to copy them directly into a filesystem store (`STORAGE_TYPE=filesystem`). Existing files are left untouched.

Set the storage environment variables to match the config that the server will load, and optionally point the entrypoint at a specific YAML file using `MOCKGRID_CONFIG`:

```bash
docker run --rm \
  -v "$PWD/initdb":/docker-entrypoint-initdb.d \
  -v "$PWD/config.yaml":/etc/mockgrid/config.yaml \
  -e STORAGE_TYPE=sqlite \
  -e STORAGE_PATH=/data/messages.db \
  -e MOCKGRID_CONFIG=/etc/mockgrid/config.yaml \
  mockgrid:latest serve
```

With this setup you can drop SQL that creates tables or INSERTs, shell scripts that prepare fixtures, and JSON payloads that represent existing messages without needing to rebuild the image. The entrypoint automatically creates the target directories and runs the scripts once before launching `mockgrid serve`.

## Configuration

Configuration is loaded from three sources (in order of precedence):
1. **Environment variables** (highest priority)
2. **Configuration file** (via `--config` or `-c`  flag)
3. **CLI flags** (override file and env vars)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_SERVER` | SMTP server hostname | `localhost` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USER` | SMTP authentication username | (optional) |
| `SMTP_PASS` | SMTP authentication password | (optional) |
| `MOCKGRID_HOST` | Host to bind the mockgrid server | `0.0.0.0` |
| `MOCKGRID_PORT` | Port to bind the mockgrid server | `5900` |
| `TEMPLATES_MODE` | Template mode: `local`, `sendgrid`, or `besteffort` | (optional) |
| `TEMPLATES_DIRECTORY` | Local templates directory | (optional) |
| `TEMPLATES_SG_KEY` | SendGrid API key for remote templates | (optional) |
| `ATTACHMENTS_DIR` | Directory to store email attachments | (optional) |
| `SENDGRID_KEY` | SendGrid API key for authentication | (optional) |
| `STORAGE_TYPE` | Storage type: `none`, `sqlite`, or `filesystem` | `none` |
| `STORAGE_PATH` | Storage path (SQLite DB file or filesystem directory) | (optional) |

### CLI Flags

Run `mockgrid serve --help` to see all available flags:

```
--config, -c <path>                 Path to configuration file
--smtp-server <hostname>            SMTP server hostname
--smtp-port <port>                  SMTP server port
--smtp-user <username>              SMTP authentication username
--smtp-pass <password>              SMTP authentication password
--mockgrid-host <host>              Host to bind on
--mockgrid-port <port>              Port to bind on
--templates-mode <mode>             Template mode (local|sendgrid|besteffort)
--templates-directory <path>        Local templates directory
--templates-key <key>               Templates API key
--attachments-dir <path>            Attachment storage directory
--sendgrid-key <key>                SendGrid API key
--storage-type <type>               Storage type (none|sqlite|filesystem)
--storage-path <path>               Storage path
```

### Configuration File (YAML)

Create a config file and pass it via `--config` or environment:

```yaml
# SMTP configuration
smtp_server: localhost
smtp_port: 587
smtp_user: ""      # Optional
smtp_pass: ""      # Optional

# Mockgrid server binding
mockgrid_host: 0.0.0.0
mockgrid_port: 5900

# Template configuration
templates:
  mode: besteffort      # local, sendgrid, or besteffort
  directory: ./templates
  template_key: ""      # SendGrid API key for remote templates

# Attachment handling
attachments:
  dir: ./attachments

# Authentication
auth:
  sendgrid_key: ""      # Optional API key for /v3/mail/send authentication
  smtp_user: ""         # Optional SMTP auth username
  smtp_pass: ""         # Optional SMTP auth password

# Message persistence
storage:
  type: none            # none, sqlite, or filesystem
  path: ""              # DB file for sqlite, directory for filesystem
```

### Configuration Precedence

Values are merged in this order (later values override earlier):
1. Environment variables
2. Configuration file values
3. CLI flag values
4. Built-in defaults

Example: If `SMTP_SERVER=prod.smtp.com` is set as an env var, but `smtp_server: localhost` is in the config file, and `--smtp-server=test.local` is passed as a flag, the flag value (`test.local`) will be used.
- Bug reports and PRs welcome. Please open issues for design discussions before large changes.

# License
This project is published under the terms in the `LICENSE` file.
