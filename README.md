# Jira Worklogger

A command-line tool for logging work hours to Jira Cloud/Server projects efficiently.

## Features

- Simple command-line interface (no GUI)
- Category aliases for quick time logging
- Automatic suggestion of epics you're working on
- Configurable time entry formats (1h30m, 1.5h, 1:30, etc.)
- Multiple configuration options (environment variables or YAML config)
- Support for both Jira Cloud and Jira Server

## Installation

### Download Pre-built Binary

Download the latest release from the [GitHub Releases page](https://github.com/liam-witterick/jira-worklogger/releases).

#### Linux/macOS

```bash
# Download the appropriate binary for your system
curl -LO https://github.com/liam-witterick/jira-worklogger/releases/download/v1.0.0/jira-worklogger-linux-amd64
chmod +x jira-worklogger-linux-amd64

# Option 1: Move to your bin directory
mv jira-worklogger-linux-amd64 ~/bin/jira-worklogger

# Option 2: System-wide installation (requires root)
sudo mv jira-worklogger-linux-amd64 /usr/local/bin/jira-worklogger
```

#### Windows

Download `jira-worklogger-windows-amd64.exe` from the releases page and rename it to `jira-worklogger.exe`. Place it in a directory that's in your PATH.

### Using the Installation Script

```bash
# Download the installation package
curl -LO https://github.com/liam-witterick/jira-worklogger/releases/download/v1.0.0/jira-worklogger-1.0.0.tar.gz

# Extract the package
tar -xzvf jira-worklogger-1.0.0.tar.gz

# Run the installation script
./install.sh
```

### Build from Source

If you prefer to build from source, you need Go 1.19 or later installed:

```bash
git clone https://github.com/liam-witterick/jira-worklogger.git
cd jira-worklogger
go build
```

## Configuration

### Create Configuration File

Create a `worklog_config.yaml` file in one of these locations:

- Current directory
- Your home directory (`~/.worklog_config.yaml`)
- System-wide (`/etc/jira-worklogger/worklog_config.yaml`)

Example configuration:

```yaml
# Minimal config for Jira Worklogger
jira_base_url: "https://yourcompany.atlassian.net"
jira_email: "your.email@company.com"
jira_api_token: "your_jira_api_token"
timezone: "Europe/London"
api_version: "3"  # Use "3" for Jira Cloud, "2" for Jira Server
log_level: "info"  # Options: debug, info, warn, error

defaults:
  category_aliases:
    meetings: "PROJ-123"
    support: "PROJ-456"
    docs: "PROJ-789"
```

### Environment Variables

You can also configure via environment variables:

```bash
export JIRA_BASE_URL="https://yourcompany.atlassian.net"
export JIRA_EMAIL="your.email@company.com"
export JIRA_API_TOKEN="your_jira_api_token"
export TIMEZONE="Europe/London"
export JIRA_API_VERSION="3"
export LOG_LEVEL="info"
export WORKLOG_CONFIG="/path/to/your/config.yaml"  # Optional, to specify config location
```

## Usage

### Basic Usage

Simply run the command:

```bash
jira-worklogger
```

You'll be prompted to enter:
1. Date (defaults to today)
2. Time entries in various formats

### Time Entry Formats

You can log time in multiple formats:

```
# Format: category=time
meetings=1h30m; support=45m; docs=2h

# Category aliases (as defined in config)
meetings=1h; PROJ-123=30m

# Multiple lines
meetings=1h
support=30m
PROJ-123=2h30m
```

Time formats supported:
- `1h30m` (1 hour, 30 minutes)
- `1.5h` (1.5 hours)
- `90m` (90 minutes)
- `1:30` (1 hour, 30 minutes)

### Command-line Options

```
jira-worklogger --help
```

The tool supports the following command-line options:

- `--help`, `-h`: Show help information
- `--version`, `-v`: Show version information
- `--date DATE`: Set the worklog date (format: YYYY-MM-DD, default: today)
- `--entries ENTRIES`: Specify time entries directly (e.g., "meetings=1h;support=30m")

#### Non-interactive Mode

You can run jira-worklogger in non-interactive mode by providing the `--entries` parameter:

```bash
# Log time for today
jira-worklogger --entries "meetings=1h;support=30m;PROJ-123=45m"

# Log time for a specific date
jira-worklogger --date "2025-09-08" --entries "meetings=2h;docs=1h30m"
```

This is useful for when you want to quickly log time without going through the interactive prompts.