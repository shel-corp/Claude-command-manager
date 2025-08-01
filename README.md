# Claude Command Manager

A standalone command library manager that allows users to enable, disable, and manage custom Claude commands from a centralized library. This tool provides both an interactive TUI interface built with Go and a traditional shell script interface.

## Overview

The command library provides an interactive CLI tool to manage custom Claude commands stored in this repository. Commands are stored as Markdown files with YAML frontmatter and can be selectively enabled by creating symlinks to `~/.claude/commands`.

## Usage

The command library automatically finds the nearest `.claude` directory by traversing up parent directories from your current location, similar to how Git finds `.git` directories. This means you can run the command from anywhere within a project that contains a `.claude` folder.

### Interactive Mode (Recommended)

**Go TUI Version:**
```bash
go run cmd/main.go
# or build and run:
go build -o command_library cmd/main.go
./command_library
```

**Shell Script Version:**
```bash
./command_library.sh
```

Interactive mode provides a full-screen interface with:
- Arrow key navigation (↑/↓) or vim-style (k/j)
- Visual highlighting of current selection
- Single-key commands for all operations
- Immediate save of all changes
- Clean and responsive interface

**Note**: Interactive mode requires a terminal environment. If run in a non-interactive environment (like CI/CD or scripts), it will display the current command list and suggest using the CLI interface instead.

### Command Line Interface

**Go Version:**
```bash
go run cmd/main.go list                     # List all available commands
go run cmd/main.go status                   # Show command status
go run cmd/main.go enable <command_name>    # Enable a specific command
go run cmd/main.go disable <command_name>   # Disable a specific command
go run cmd/main.go rename <cmd> <new_name>  # Rename a command
go run cmd/main.go help                     # Show help
```

**Shell Script Version:**
```bash
./command_library.sh list                     # List all available commands
./command_library.sh status                   # Show command status
./command_library.sh enable <command_name>    # Enable a specific command
./command_library.sh disable <command_name>   # Disable a specific command
./command_library.sh rename <cmd> <new_name>  # Rename a command
./command_library.sh help                     # Show help
```

## Features

- **Directory Traversal**: Automatically finds the nearest `.claude` directory from your current location
- **Interactive Interface**: Professional TUI with arrow keys and single-key actions
- **Immediate Save**: All changes are saved immediately, no session tracking needed
- **Symlink Management**: Automatic creation/removal of symlinks to `~/.claude/commands`
- **Command Renaming**: Rename commands without affecting source files
- **Status Tracking**: JSON configuration tracks enabled/disabled state and renames
- **Error Handling**: Comprehensive error handling and broken symlink cleanup
- **YAML Parsing**: Extracts descriptions from command file frontmatter
- **Cross-Directory Support**: Works from any subdirectory within a `.claude`-enabled project

## File Structure

```
claude_command_manager/
├── README.md                     # This file
├── go.mod                        # Go module definition
├── go.sum                        # Go dependency checksums
├── cmd/
│   └── main.go                   # Go TUI application entry point
├── internal/
│   ├── commands/
│   │   └── commands.go           # Command management logic
│   ├── config/
│   │   ├── config.go             # Configuration management
│   │   └── paths.go              # Path resolution
│   └── tui/                      # Terminal UI components
│       ├── model.go              # Bubble Tea model
│       ├── view.go               # UI rendering
│       ├── update.go             # Event handling
│       └── styles.go             # UI styling
├── command_library.sh            # Legacy shell script interface
├── commands/                     # Example command files
│   ├── analyze_code.md
│   ├── debug_helper.md
│   └── project_summary.md
└── .config.json                  # Configuration tracking (auto-generated)
```

## Command File Format

Command files are Markdown files with YAML frontmatter:

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

## Configuration

Configuration is stored in `.config.json`:

```json
{
  "commands": {
    "command_name": {
      "enabled": true,
      "original_name": "command_name",
      "display_name": "renamed_command",
      "source_path": ".claude/command_library/commands/command_name.md"
    }
  }
}
```

## Dependencies

**For Go TUI Version:**
- Go 1.24.3 or later

**For Shell Script Version:**
- `jq` - For JSON configuration management

Install via Homebrew:
```bash
brew install jq
```

## Installation

1. Clone or download this repository
2. For Go version, ensure Go is installed
3. For shell script version, ensure `jq` is installed
4. Place the command manager in a convenient location
5. Add your custom command files to the `commands/` directory

## Building (Go Version)

Use the included build script:

```bash
# Make build script executable
chmod +x build.sh

# Build for current platform
./build.sh build

# Build for all platforms
./build.sh all

# Run tests
./build.sh test

# Clean build artifacts
./build.sh clean
```

Or build manually:
```bash
go build -o command_library cmd/main.go
```

## Error Handling

The tool includes comprehensive error handling for:
- Missing dependencies
- Permission issues
- Broken symlinks
- Invalid JSON configuration
- Missing directories
- Corrupted YAML frontmatter

## Example Workflow

1. **Add New Commands**: Place `.md` files in the `commands/` directory 
2. **Launch Manager**: 
   - Go version: `go run cmd/main.go` or build and run the binary
   - Shell version: `./command_library.sh`
3. **Navigate**: Use ↑/↓ arrow keys or k/j to select commands (Go version)
4. **Make Changes**:
   - **Go version**: Changes are saved immediately
     - Press Enter to toggle enabled/disabled
     - Press 'r' to rename a command
     - Press 'd' to disable a command
     - Press 'l' to toggle symlink location (user vs project)
   - **Shell version**: Session-based with explicit save
5. **Exit**: Press 'q' to quit

## Troubleshooting

- **"No .claude directory found"**: Make sure you're running the command from within a directory that contains a `.claude` folder, or any of its subdirectories
- **Broken symlinks**: The tool automatically cleans up broken symlinks on startup
- **Configuration corruption**: Invalid JSON is automatically backed up and reset
- **Permission issues**: Ensure write access to `~/.claude/commands` directory

## Architecture

**Go TUI Version** uses:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- [Bubbles](https://github.com/charmbracelet/bubbles) for UI components

**Shell Script Version** uses:
- Pure shell scripting with `jq` for JSON configuration management
- Interactive menu system for command management

Both versions provide the same core functionality with different user interfaces and can be used interchangeably. The Go version offers a more modern TUI experience, while the shell version provides traditional menu-based interaction.