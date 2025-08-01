# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-08-01

### Added
- Initial release of Claude Command Manager
- Interactive TUI interface built with Bubble Tea
- CLI mode for scripting and automation
- Command management features:
  - Enable/disable commands
  - Rename commands with display names
  - Toggle symlink locations (user vs project)
- Automatic directory traversal to find `.claude` directories
- Comprehensive error handling and broken symlink cleanup
- JSON configuration management
- Cross-platform build support
- Homebrew formula for easy installation

### Features
- **Interactive Mode**: Full-screen TUI with arrow key navigation
- **CLI Mode**: Command-line interface for automation
- **Directory Detection**: Automatically finds nearest `.claude` directory
- **Symlink Management**: Creates symlinks to both user and project command directories
- **Configuration Tracking**: JSON-based state management
- **Immediate Save**: All changes saved instantly, no session management needed

[Unreleased]: https://github.com/sheltontolbert/claude_command_manager/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/sheltontolbert/claude_command_manager/releases/tag/v1.0.0