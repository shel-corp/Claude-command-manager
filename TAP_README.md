# Homebrew Tap for Claude Tools

This repository serves as a Homebrew tap for Claude Code related tools, starting with the Claude Command Manager.

## Installation

### Quick Installation

```bash
# Add the tap
brew tap sheltontolbert/claude-tools

# Install ccm
brew install ccm
```

### Alternative: Direct Installation

```bash
# Install directly without adding tap permanently
brew install sheltontolbert/claude-tools/ccm
```

## Available Formulas

### ccm (Claude Command Manager)

A TUI (Terminal User Interface) application for managing Claude Code custom commands.

**Features:**
- Interactive terminal interface built with Bubble Tea
- CLI mode for scripting and automation
- Automatic directory traversal to find `.claude` directories
- Command enable/disable management
- Command renaming with display names
- Symlink location switching (user vs project)
- Immediate configuration saving

**Usage:**
```bash
# Launch interactive TUI
ccm

# CLI commands
ccm list                     # List all commands
ccm status                   # Show command status
ccm enable <command_name>    # Enable a command
ccm disable <command_name>   # Disable a command
ccm rename <cmd> <new_name>  # Rename a command
ccm help                     # Show help
```

## Requirements

- **macOS 10.15+** or **Linux**
- **Go 1.24+** (for building from source)
- **jq** (optional, for shell script compatibility)

## Development

This tap is automatically maintained using GoReleaser and GitHub Actions. When new releases are tagged in the source repositories, the formulas are automatically updated.

### Manual Formula Updates

If you need to manually update a formula:

1. Fork this repository
2. Update the formula in the `Formula/` directory
3. Test the formula locally:
   ```bash
   brew install --build-from-source ./Formula/ccm.rb
   brew test ccm
   brew audit --strict --online ccm
   ```
4. Submit a pull request

## Troubleshooting

### Formula Issues

```bash
# Clean up and retry
brew cleanup
brew doctor

# Reinstall if needed
brew uninstall ccm
brew install ccm
```

### Build Issues

```bash
# Install dependencies
brew install go

# For shell script compatibility
brew install jq
```

### Reporting Issues

If you encounter issues with any formula in this tap:

1. Check if the issue exists with the upstream project
2. Report formula-specific issues in this repository
3. Report application issues in the respective tool's repository

## Links

- **Claude Command Manager**: https://github.com/shel-corp/Claude-command-manager
- **Homebrew Documentation**: https://docs.brew.sh/
- **Homebrew Formula Cookbook**: https://docs.brew.sh/Formula-Cookbook

---

*This tap is maintained by [@shel-corp](https://github.com/shel-corp)*