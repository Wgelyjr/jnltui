# Journal TUI

A terminal-based journal application with a responsive interface.

## Features

- View, create, edit, and delete journal entries
- Responsive interface that adapts to terminal size
- Configurable storage location for journal entries
- Support for system-wide and user-specific configuration

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/journal-tui.git
cd journal-tui

# Build the application
go build -o bin/journal-tui

# Install the application (optional)
sudo cp journal-tui /usr/local/bin/
```

## Configuration

Journal TUI uses a configuration file system that follows standard Unix conventions:

1. Default configuration (built-in)
2. System-wide configuration: `/etc/journal-tui/config.yaml`
3. User-specific configuration: `~/.journal-tui/config.yaml`
4. Current directory configuration: `./journal-tui.yaml` (useful for development)

Each level overrides the previous one, so user settings take precedence over system settings.

### Configuration File Format

The configuration file is in YAML format:

```yaml
entries_dir: ~/.journal-tui/entries
dev_mode: false
```

### Configuration Options

- `entries_dir`: Directory where journal entries are stored
  - Default: `~/.journal-tui/entries`
  - For development: `./entries`
- `dev_mode`: Enable development mode
  - Default: `false`
  - When `true`, more verbose logging is enabled

### Development vs. Production

- In production mode (default), entries are stored in `~/.journal-tui/entries`
- In development mode, entries are stored in `./entries` (current working directory)

To run in development mode, create a `journal-tui.yaml` file in your working directory:

```yaml
dev_mode: true
entries_dir: ./entries
```


## License

MIT
