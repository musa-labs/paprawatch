# paprawatch

`paprawatch` is a Go-based CLI tool that monitors local directories for new PDF files and automatically uploads them to a [Papra](https://papra.app) organization via the official REST API.

## Features

- **Real-time Monitoring**: Uses `fsnotify` to detect new files as soon as they are created.
- **Bulk Scanning**: A dedicated `scan` command to traverse your directories and upload existing PDFs.
- **Smart Tracking**: Uses a local SQLite database (`~/.paprawatch/paprawatch.db`) to ensure files are only uploaded once, even if they are moved or the tool is restarted.
- **Robust Uploads**: Built-in exponential backoff and retry logic for reliable uploads under heavy load.
- **Persistent Config**: Interactive setup that saves your credentials and preferences to `~/.config/paprawatch/config.yaml`.
- **Modern CLI**: Built with `urfave/cli/v3` and `charmbracelet/huh`.

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

## Getting Started

The easiest way to get started is by running the interactive setup:

```bash
./paprawatch setup
```

This will walk you through configuring your Organization ID, API Token, and the directories you want to watch.

## Usage

### 1. Watch Mode (Default)

Start the real-time watcher. It will monitor your configured directories and upload any new PDFs.

```bash
./paprawatch
```

### 2. Scan Mode

Scan your configured directories for existing PDFs and upload them. It uses file hashing to skip anything already seen by `paprawatch`.

```bash
./paprawatch scan
```

You can also override the directories to scan via the CLI:

```bash
./paprawatch scan --dir ~/Documents --dir ~/Downloads
```

### Options & Flags

| Flag | Alias | Environment Variable | Description | Default |
|------|-------|----------------------|-------------|---------|
| `--dir` | `-d` | `PAPRA_WATCH_DIR` | Directories to watch or scan (can be specified multiple times). | `.` |
| `--org` | `-o` | `PAPRA_ORG_ID` | Papra Organization ID. | - |
| `--token` | `-t` | `PAPRA_API_TOKEN` | Papra API Token. | - |
| `--url` | `-u` | `PAPRA_API_URL` | Papra instance URL. | `https://api.papra.app` |
| `--ocr` | - | `PAPRA_OCR_LANGUAGES` | OCR Languages (optional, e.g. 'eng,fra'). | - |

## How It Works

1. **Configuration**: Credentials and watch directories are stored in your home directory. CLI flags and environment variables take precedence over saved config.
2. **Deduplication**: Every PDF found is hashed (SHA-256). The hash is checked against a local SQLite database. If the hash exists, the file is skipped.
3. **Resilience**:
    - If an upload fails due to a network error or server load, the tool uses **exponential backoff** to retry up to 5 times.
    - If the server reports the document already exists (409 Conflict), the tool records this in the local database and continues.
4. **Streaming**: Files are streamed to the API using `io.Pipe` to keep memory usage low even with large PDF collections.

## Development

### Running Tests

To run the internal test suite:

```bash
go test ./...
```

### Project Structure

- `main.go`: CLI entry point and command routing.
- `api/`: Papra REST API client with retry logic.
- `config/`: Configuration management and interactive setup.
- `db/`: SQLite persistence for file tracking.
- `scanner/`: Recursive directory walking and hashing.
- `watcher/`: Real-time file system events.
