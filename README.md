# paprawatch

`paprawatch` is a Go-based CLI tool that monitors a local directory for new files and automatically uploads them to a [Papra](https://papra.app) organization via the official REST API.

## Features

- **Real-time Monitoring**: Uses `fsnotify` to detect new files as soon as they are created.
- **Automated Uploads**: Automatically constructs `multipart/form-data` requests to the Papra API.
- **Secure Authentication**: Supports API tokens via CLI flags or environment variables.
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

You can configure `paprawatch` using environment variables for a cleaner command or better security:

```bash
export PAPRA_API_TOKEN="your-api-token"
export PAPRA_ORG_ID="your-org-id"
export PAPRA_WATCH_DIR="./uploads"
./paprawatch
```

### Options

| Flag | Alias | Environment Variable | Description | Default |
|------|-------|----------------------|-------------|---------|
| `--dir` | `-d` | `PAPRA_WATCH_DIR` | **Required.** Directory to watch for new files. | - |
| `--org` | `-o` | `PAPRA_ORG_ID` | **Required.** Papra Organization ID. | - |
| `--token` | `-t` | `PAPRA_API_TOKEN` | **Required.** Papra API Token. | - |
| `--url` | `-u` | `PAPRA_API_URL` | Papra instance URL. | `https://api.papra.app` |
| `--ocr` | - | `PAPRA_OCR_LANGUAGES` | OCR Languages (optional, e.g. 'eng,fra'). | - |

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
