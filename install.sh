#!/bin/bash

# Jira Worklogger Installation Script
# This script installs the jira-worklogger binary and sets up the configuration

set -e
echo "=== Jira Worklogger Installation ==="

# Determine target directory
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/jira-worklogger"
HOME_CONFIG="$HOME/.worklog_config.yaml"

# Check if we have root privileges for system-wide installation
if [ "$EUID" -ne 0 ]; then
  echo "[notice] Running without root privileges, installing to $HOME/bin"
  INSTALL_DIR="$HOME/bin"
  mkdir -p "$INSTALL_DIR"
fi

# Copy binary
echo "[info] Installing jira-worklogger binary to $INSTALL_DIR"
cp jira-worklogger "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/jira-worklogger"

# Set up configuration
if [ "$EUID" -eq 0 ] && [ ! -d "$CONFIG_DIR" ]; then
  echo "[info] Creating system-wide config directory $CONFIG_DIR"
  mkdir -p "$CONFIG_DIR"
  
  # Only copy if it doesn't exist
  if [ ! -f "$CONFIG_DIR/worklog_config.yaml" ] && [ -f "worklog_config.yaml.example" ]; then
    echo "[info] Creating system-wide config file $CONFIG_DIR/worklog_config.yaml"
    cp worklog_config.yaml.example "$CONFIG_DIR/worklog_config.yaml"
    echo "[warn] Please update the API token in $CONFIG_DIR/worklog_config.yaml"
  fi
else
  # User-level installation
  if [ ! -f "$HOME_CONFIG" ] && [ -f "worklog_config.yaml.example" ]; then
    echo "[info] Creating user config file at $HOME_CONFIG"
    cp worklog_config.yaml.example "$HOME_CONFIG"
    echo "[warn] Please update the API token in $HOME_CONFIG"
  fi
fi

# Check if the directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "[warn] $INSTALL_DIR is not in your PATH. Add it by running:"
  echo "export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo "[success] Installation complete!"
echo ""
echo "Usage: jira-worklogger [options]"
echo "Run 'jira-worklogger --help' for more information."
