# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based TUI application that manages Claude Code custom commands. It provides both an interactive terminal interface and CLI commands for enabling/disabling/renaming Claude commands stored as Markdown files with YAML frontmatter.

## Development Commands

### Building
```bash
# Build for current platform
go build -o command_library cmd/main.go
# or use the build script
./build.sh build

# Build for all platforms
./build.sh all

# Clean build artifacts
./build.sh clean
```

### Testing
```bash
# Run all tests
go test ./...
# or use the build script
./build.sh test
```

### Running
```bash
# Run directly with Go
go run cmd/main.go

# Run built binary
./command_library

# CLI mode examples
go run cmd/main.go list
go run cmd/main.go enable <command_name>
go run cmd/main.go disable <command_name>
```

## Compilation Instructions

### Prerequisites
- Go 1.20 or higher
- Make sure GOPATH and Go binary are in your system PATH
- Required development tools: 
  - `git`
  - `make` (optional but recommended)
  - `go` compiler

### Compilation Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/shelton-tolbert/claude_command_manager.git
   cd claude_command_manager
   ```

2. Download dependencies:
   ```bash
   go mod download
   go mod tidy
   ```

3. Compile the project:
   ```bash
   # Standard compilation
   go build -o command_library cmd/main.go

   # Compile with optimization
   go build -ldflags="-s -w" -o command_library cmd/main.go

   # Cross-platform compilation
   GOOS=darwin GOARCH=amd64 go build -o command_library_mac cmd/main.go
   GOOS=linux GOARCH=amd64 go build -o command_library_linux cmd/main.go
   GOOS=windows GOARCH=amd64 go build -o command_library.exe cmd/main.go
   ```

4. Optional: Install using build script
   ```bash
   ./build.sh build   # Builds for current platform
   ./build.sh all     # Builds for all supported platforms
   ```

### Troubleshooting Compilation
- Ensure all dependencies are downloaded: `go mod download`
- Check Go version compatibility: `go version`
- Verify GOPATH is correctly set: `echo $GOPATH`

## Architecture

### Core Components

- **cmd/main.go**: Application entry point with CLI argument handling and TUI initialization
- **internal/tui/**: Bubble Tea TUI implementation
  - `model.go`: Application state and data structures
  - `view.go`: UI rendering logic
  - `update.go`: Event handling and state transitions
  - `styles.go`: UI styling definitions
- **internal/commands/**: Command management logic (scanning, enabling/disabling, symlink management)
- **internal/config/**: Configuration management (JSON config file, path resolution)

### Key Architecture Patterns

1. **Directory Traversal**: Automatically finds `.claude` directory by traversing up parent directories (like Git)
2. **Dual Interface**: Supports both interactive TUI (default) and CLI mode for scripting
3. **Immediate Save**: All changes are saved immediately to JSON config file, no session tracking
4. **Symlink Management**: Creates/removes symlinks to either `~/.claude/commands/` (user) or `<project>/.claude/commands/` (project)
5. **State Management**: Uses Bubble Tea's Elm-like architecture with messages and updates

### Dependencies

- **Bubble Tea**: TUI framework for interactive terminal applications
- **Lipgloss**: Styling library for terminal UIs
- **Bubbles**: Pre-built UI components (list, text input)

### Configuration

- Commands are stored as `.md` files in `commands/` directory with YAML frontmatter
- Configuration is tracked in `.config.json` with command state (enabled/disabled, renames, symlink locations)
- Supports two symlink destinations: user-wide (`~/.claude/commands/`) and project-specific (`<project>/.claude/commands/`)

### Command File Format

Commands are Markdown files with YAML frontmatter:
```markdown
---
description: Brief description of what the command does
allowed-tools: Read(*), Bash(git:*)  # Optional
---

# Command Content
Your command prompt goes here.
$ARGUMENTS
Additional instructions...
```

### Error Handling

- Comprehensive error handling for missing dependencies, permission issues, broken symlinks
- Automatic cleanup of broken symlinks on startup
- JSON configuration corruption detection and recovery

## Homebrew Features

- Use `brew install` to install the command library
- Homebrew tap is available at `shelton-tolbert/claude`
- Install with: `brew tap shelton-tolbert/claude && brew install command_library`
- Update Homebrew package by running `brew upgrade command_library`
- Check current Homebrew package version with `brew info command_library`
```