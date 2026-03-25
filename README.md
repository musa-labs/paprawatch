# paprawatch

`paprawatch` is a Go-based CLI tool that monitors a local directory for new files and automatically uploads them to a [Papra](https://papra.app) organization via the official REST API.

## Features

- **Real-time Monitoring**: Uses `fsnotify` to detect new files as soon as they are created.
- **Automated Uploads**: Automatically constructs `multipart/form-data` requests to the Papra API.
- **Secure Authentication**: Supports API tokens via CLI flags or the `PAPRA_API_TOKEN` environment variable.
- **Modern CLI**: Built with `urfave/cli/v3`.

## Prerequisites

- [Go](https://go.dev/doc/install) 1.21 or higher.
- A Papra account and an [API Token](https://api.papra.app/api/api-keys/current) with `documents:create` permissions.
- Your Papra Organization ID.

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/musa-labs/paprawatch.git
cd paprawatch
go build -o paprawatch
```

## Usage

You can start watching a directory by providing the directory path, organization ID, and your API token.

### Using CLI Flags

```bash
./paprawatch --dir ./uploads --org "your-org-id" --token "your-api-token"
```

### Using Environment Variables

For better security, you can export your API token as an environment variable:

```bash
export PAPRA_API_TOKEN="your-api-token"
./paprawatch -d ./uploads -o "your-org-id"
```

### Options

| Flag | Alias | Environment Variable | Description | Default |
|------|-------|----------------------|-------------|---------|
| `--dir` | `-d` | - | **Required.** Directory to watch for new files. | - |
| `--org` | `-o` | - | **Required.** Papra Organization ID. | - |
| `--token` | `-t` | `PAPRA_API_TOKEN` | **Required.** Papra API Token. | - |
| `--url` | `-u` | - | Papra instance URL. | `https://api.papra.app` |

## How It Works

1. **Initialization**: The tool parses your configuration and initializes an API client and a file system watcher.
2. **Watching**: It uses the `fsnotify` library to hook into operating system events. It specifically listens for `Create` events in the target directory.
3. **Trigger**: When a new file is detected, its path is passed to the upload handler.
4. **Upload**: The tool opens the file, wraps it in a multipart form, adds the `Authorization: Bearer <token>` header, and sends a `POST` request to `/api/organizations/:orgId/documents`.
5. **Logging**: Success or failure of each upload is logged directly to the terminal.

## Development

### Running Tests

To run the internal test suite:

```bash
go test ./...
```

### Project Structure

- `main.go`: Entry point and CLI flag definitions.
- `api/client.go`: Handles HTTP communication with Papra.
- `watcher/watcher.go`: Manages the `fsnotify` lifecycle.
