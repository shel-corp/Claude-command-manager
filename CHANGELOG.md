# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2025-08-04

### Added
- **🎨 Complete Theme System**: Professional theme management with 6 built-in themes
  - Default (Classic blue theme)
  - Monochrome (Elegant grayscale)
  - Solarized (Scientifically-designed color palette)
  - Dracula (Dark theme with vibrant accents)
  - Nord (Arctic-inspired cool blues)
  - **Gruvbox Material** (Warm, earthy theme designed to protect developers' eyes)
- **⚙️ Settings Menu**: Comprehensive settings system accessible from main menu
  - Interactive theme picker with live color previews
  - Keyboard navigation (↑/↓ to browse, Enter to apply)
  - Current theme indicator with checkmark
- **🌈 Adaptive Color Support**: Themes automatically adjust for light/dark terminals
- **💾 Theme Persistence**: Theme choices save automatically across sessions
- **🔄 Remote Repository Features**: Enhanced command import and browsing
  - Repository browsing with category filtering
  - Command preview functionality
  - Batch import capabilities
  - GitHub URL parsing and validation
- **📦 Improved Caching System**: Enhanced performance with intelligent caching
- **🛠️ Enhanced UI Components**: Polished interface with better error handling

### Enhanced
- **Terminal UI**: Significantly improved user interface with theme support
- **Configuration Management**: Thread-safe theme configuration with JSON persistence
- **Error Handling**: Comprehensive error handling with user-friendly messages
- **Performance**: Optimized style generation and caching for faster theme switching
- **Navigation**: Enhanced keyboard navigation throughout the application

### Technical Improvements
- Added comprehensive theme architecture with `internal/theme/` package
- Implemented thread-safe theme manager with mutex protection
- Created adaptive color system using lipgloss.AdaptiveColor
- Enhanced TUI state management for settings and theme navigation
- Improved code organization with better separation of concerns

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

[Unreleased]: https://github.com/shel-corp/Claude-command-manager/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/shel-corp/Claude-command-manager/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/shel-corp/Claude-command-manager/releases/tag/v1.0.0