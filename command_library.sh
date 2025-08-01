#!/usr/bin/env zsh
# shellcheck shell=bash

set -e

# Get script directory and project root
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMMAND_LIBRARY_DIR="$SCRIPT_DIR"
CONFIG_FILE="$COMMAND_LIBRARY_DIR/.config.json"
COMMANDS_DIR="$COMMAND_LIBRARY_DIR/commands"
CLAUDE_COMMANDS_DIR="$HOME/.claude/commands"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${RESET} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${RESET} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${RESET} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${RESET} $1"
}

# Initialize configuration file if it doesn't exist
init_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        print_info "Initializing configuration file..."
        echo '{"commands": {}}' | jq . > "$CONFIG_FILE"
        print_success "Configuration file created at $CONFIG_FILE"
    fi
}

# Load configuration from JSON file
load_config() {
    init_config
    if ! cat "$CONFIG_FILE" | jq empty > /dev/null 2>&1; then
        print_error "Invalid JSON in config file. Reinitializing..."
        echo '{"commands": {}}' | jq . > "$CONFIG_FILE"
    fi
}

# Save configuration to JSON file
save_config() {
    local temp_file=$(mktemp)
    echo "$1" | jq . > "$temp_file"
    if jq empty "$temp_file" > /dev/null 2>&1; then
        mv "$temp_file" "$CONFIG_FILE"
        print_success "Configuration saved"
    else
        print_error "Invalid JSON configuration. Not saving."
        rm "$temp_file"
        return 1
    fi
}

# Get command status from config
get_command_status() {
    local command_name="$1"
    local config=$(cat "$CONFIG_FILE")
    
    # Check if command exists in config
    if echo "$config" | jq -e ".commands[\"$command_name\"]" > /dev/null 2>&1; then
        echo "$config" | jq -r ".commands[\"$command_name\"].enabled // false"
    else
        echo "false"
    fi
}

# Get command display name from config (handles renames)
get_command_display_name() {
    local command_name="$1"
    local config=$(cat "$CONFIG_FILE")
    
    if echo "$config" | jq -e ".commands[\"$command_name\"]" > /dev/null 2>&1; then
        echo "$config" | jq -r ".commands[\"$command_name\"].display_name // \"$command_name\""
    else
        echo "$command_name"
    fi
}

# Update command in config
update_command_config() {
    local command_name="$1"
    local enabled="$2"
    local display_name="$3"
    local source_path="$4"
    
    local config=$(cat "$CONFIG_FILE")
    
    # Update or create command entry
    config=$(echo "$config" | jq --arg name "$command_name" \
                                 --arg enabled "$enabled" \
                                 --arg display "$display_name" \
                                 --arg source "$source_path" \
                                 '.commands[$name] = {
                                     "enabled": ($enabled == "true"),
                                     "original_name": $name,
                                     "display_name": $display,
                                     "source_path": $source
                                 }')
    
    save_config "$config"
}

# Remove command from config
remove_command_config() {
    local command_name="$1"
    local config=$(cat "$CONFIG_FILE")
    
    config=$(echo "$config" | jq --arg name "$command_name" 'del(.commands[$name])')
    save_config "$config"
}

# Print usage information
print_usage() {
    echo -e "Claude Command Manager (ccm)"
    echo -e "${YELLOW}Usage:${RESET}"
    echo -e "  ./bin/command_library                          Launch interactive command manager"
    echo -e "  ./bin/command_library list                     List all available commands"
    echo -e "  ./bin/command_library status                   Show current command status"
    echo -e "  ./bin/command_library enable <command_name>    Enable a specific command"
    echo -e "  ./bin/command_library disable <command_name>   Disable a specific command"
    echo -e "  ./bin/command_library rename <cmd> <new_name>  Rename a command"
    echo -e "  ./bin/command_library help                     Show this help message"
    echo
    echo -e "${YELLOW}Interactive Mode Controls:${RESET}"
    echo -e "  ↑/↓ or k/j   Navigate up/down"
    echo -e "  Enter        Toggle command enabled/disabled"
    echo -e "  r            Rename selected command"
    echo -e "  s            Save all changes and exit"
    echo -e "  q            Quit (with unsaved changes prompt)"
}

# Main function
main() {
    case "${1:-interactive}" in
        "help"|"-h"|"--help")
            print_usage
            ;;
        "list")
            list_commands
            ;;
        "status")
            show_status
            ;;
        "enable")
            if [[ -z "$2" ]]; then
                print_error "Usage: ./bin/command_library enable <command_name>"
                exit 1
            fi
            enable_command "$2"
            ;;
        "disable")
            if [[ -z "$2" ]]; then
                print_error "Usage: ./bin/command_library disable <command_name>"
                exit 1
            fi
            disable_command "$2"
            ;;
        "rename")
            if [[ -z "$2" ]] || [[ -z "$3" ]]; then
                print_error "Usage: ./bin/command_library rename <command_name> <new_name>"
                exit 1
            fi
            rename_command "$2" "$3"
            ;;
        "interactive"|"")
            interactive_mode
            ;;
        *)
            print_error "Unknown command: $1"
            print_usage
            exit 1
            ;;
    esac
}

# Parse YAML frontmatter from a markdown file
parse_yaml_frontmatter() {
    local file_path="$1"
    local key="$2"
    
    if [[ ! -f "$file_path" ]]; then
        return 1
    fi
    
    # Extract YAML frontmatter between --- delimiters
    local yaml_content=$(sed -n '1,/^---$/p' "$file_path" | sed '1d;$d' 2>/dev/null)
    
    if [[ -z "$yaml_content" ]]; then
        return 1
    fi
    
    # Simple YAML parsing for key: value pairs
    echo "$yaml_content" | grep "^$key:" | sed "s/^$key:[[:space:]]*//" | sed 's/^["'\'']//' | sed 's/["'\'']$//'
}

# Get description from command file
get_command_description() {
    local file_path="$1"
    local description=$(parse_yaml_frontmatter "$file_path" "description")
    
    if [[ -n "$description" ]]; then
        echo "$description"
    else
        echo "No description available"
    fi
}

# Scan commands directory for .md files
scan_commands() {
    if [[ ! -d "$COMMANDS_DIR" ]]; then
        print_warning "Commands directory not found: $COMMANDS_DIR"
        return 1
    fi
    
    # Find all .md files and extract basenames, then sort
    find "$COMMANDS_DIR" -name "*.md" -type f 2>/dev/null | \
        while IFS= read -r file; do
            basename "$file" .md
        done | sort
}

# List all available commands with their status
list_commands() {
    local commands=($(scan_commands))
    
    if [[ ${#commands[@]} -eq 0 ]]; then
        print_warning "No command files found in $COMMANDS_DIR"
        return 0
    fi
    
    for command in "${commands[@]}"; do
        local file_path="$COMMANDS_DIR/$command.md"
        local description=$(get_command_description "$file_path")
        local enabled=$(get_command_status "$command")
        local display_name=$(get_command_display_name "$command")
        
        local status_symbol="[ ]"
        if [[ "$enabled" == "true" ]]; then
            status_symbol="${GREEN}[X]${RESET}"
        else
            status_symbol="[ ]"
        fi
        
        echo -e "$status_symbol ${BLUE}$display_name${RESET}: $description"
    done
}

# Show status of all commands
show_status() {
    local config=$(cat "$CONFIG_FILE")
    local enabled_count=0
    local total_count=0
    local commands=($(scan_commands))
    
    if [[ ${#commands[@]} -eq 0 ]]; then
        print_warning "No command files found"
        return 0
    fi
    
    for command in "${commands[@]}"; do
        local enabled=$(get_command_status "$command")
        local display_name=$(get_command_display_name "$command")
        
        if [[ "$enabled" == "true" ]]; then
            echo -e "${GREEN}✓${RESET} $display_name (enabled)"
            ((enabled_count++))
        else
            echo -e "${YELLOW}○${RESET} $display_name (disabled)"
        fi
        ((total_count++))
    done
    
    echo
    echo "Summary: $enabled_count/$total_count commands enabled"
    
    # Show symlink status
    echo
    if [[ -d "$CLAUDE_COMMANDS_DIR" ]]; then
        local symlink_count=$(find "$CLAUDE_COMMANDS_DIR" -type l | wc -l | tr -d ' ')
        echo "Symlink status: $symlink_count symlinks in $CLAUDE_COMMANDS_DIR"
    else
        print_warning "Claude commands directory not found: $CLAUDE_COMMANDS_DIR"
    fi
}

# Create symlink for a command
create_command_symlink() {
    local command_name="$1"
    local display_name="$2"
    local source_file="$COMMANDS_DIR/$command_name.md"
    local target_file="$CLAUDE_COMMANDS_DIR/$display_name.md"
    
    if [[ ! -f "$source_file" ]]; then
        print_error "Source command file not found: $source_file"
        return 1
    fi
    
    # Check if target already exists
    if [[ -e "$target_file" ]]; then
        if [[ -L "$target_file" ]]; then
            # It's a symlink, check if it points to our file
            local current_target=$(readlink "$target_file")
            if [[ "$current_target" == "$source_file" ]]; then
                print_info "Symlink already exists and is correct: $display_name"
                return 0
            else
                print_warning "Symlink exists but points to different file: $target_file -> $current_target"
                return 1
            fi
        else
            print_error "File already exists (not a symlink): $target_file"
            return 1
        fi
    fi
    
    # Create the symlink with absolute path
    local absolute_source=$(realpath "$source_file")
    if ln -s "$absolute_source" "$target_file"; then
        print_success "Created symlink: $display_name -> $command_name"
        return 0
    else
        print_error "Failed to create symlink: $target_file"
        return 1
    fi
}

# Remove symlink for a command
remove_command_symlink() {
    local display_name="$1"
    local target_file="$CLAUDE_COMMANDS_DIR/$display_name.md"
    
    if [[ ! -e "$target_file" ]]; then
        print_info "Symlink does not exist: $display_name"
        return 0
    fi
    
    if [[ -L "$target_file" ]]; then
        if rm "$target_file"; then
            print_success "Removed symlink: $display_name"
            return 0
        else
            print_error "Failed to remove symlink: $target_file"
            return 1
        fi
    else
        print_error "File is not a symlink, not removing: $target_file"
        return 1
    fi
}

# Enable a command (create symlink and update config)
enable_command() {
    local command_name="$1"
    local display_name="${2:-$command_name}"
    local source_path="$COMMANDS_DIR/$command_name.md"
    
    if [[ ! -f "$source_path" ]]; then
        print_error "Command file not found: $source_path"
        return 1
    fi
    
    if create_command_symlink "$command_name" "$display_name"; then
        update_command_config "$command_name" "true" "$display_name" "$source_path"
        print_success "Enabled command: $display_name"
        return 0
    else
        return 1
    fi
}

# Disable a command (remove symlink and update config)
disable_command() {
    local command_name="$1"
    local display_name=$(get_command_display_name "$command_name")
    local source_path="$COMMANDS_DIR/$command_name.md"
    
    if remove_command_symlink "$display_name"; then
        update_command_config "$command_name" "false" "$display_name" "$source_path"
        print_success "Disabled command: $display_name"
        return 0
    else
        return 1
    fi
}

# Rename a command (update symlink and config)
rename_command() {
    local command_name="$1"
    local new_display_name="$2"
    local source_path="$COMMANDS_DIR/$command_name.md"
    
    # Check if command file exists
    if [[ ! -f "$source_path" ]]; then
        print_error "Command file not found: $source_path"
        return 1
    fi
    
    local old_display_name=$(get_command_display_name "$command_name")
    local enabled=$(get_command_status "$command_name")
    
    if [[ "$old_display_name" == "$new_display_name" ]]; then
        print_info "Name unchanged: $new_display_name"
        return 0
    fi
    
    # If command is enabled, update the symlink
    if [[ "$enabled" == "true" ]]; then
        # Remove old symlink
        if ! remove_command_symlink "$old_display_name"; then
            print_error "Failed to remove old symlink: $old_display_name"
            return 1
        fi
        
        # Create new symlink
        if ! create_command_symlink "$command_name" "$new_display_name"; then
            print_error "Failed to create new symlink: $new_display_name"
            # Try to restore old symlink
            create_command_symlink "$command_name" "$old_display_name"
            return 1
        fi
    fi
    
    # Update config
    update_command_config "$command_name" "$enabled" "$new_display_name" "$source_path"
    print_success "Renamed command: $old_display_name -> $new_display_name"
    return 0
}

# Toggle command enabled/disabled status
toggle_command() {
    local command_name="$1"
    local current_status=$(get_command_status "$command_name")
    
    if [[ "$current_status" == "true" ]]; then
        disable_command "$command_name"
    else
        enable_command "$command_name"
    fi
}


# Display the interactive menu
display_menu() {
    # Use global commands array directly
    local total_commands=${#commands[@]}
    
    clear 2>/dev/null || echo "===================="
    echo "Claude Command Manager (ccm)"
    echo
    
    if [[ $total_commands -eq 0 ]]; then
        echo "No commands found."
        return
    fi
    
    echo "Current Commands:"
    local i=1
    for command in "${commands[@]}"; do
        local file_path="$COMMANDS_DIR/$command.md"
        local description=$(get_command_description "$file_path")
        local enabled=$(get_command_status "$command")
        local display_name=$(get_command_display_name "$command")
        
        local status_symbol="[ ]"
        if [[ "$enabled" == "true" ]]; then
            status_symbol="${GREEN}[X]${RESET}"
        fi
        
        echo -e "$i. $status_symbol ${BLUE}$display_name${RESET}: $description"
        ((i++))
    done
    
    # Show session changes if any (session_changes is a global array)
    if [[ ${#session_changes[@]} -gt 0 ]]; then
        echo
        echo "${YELLOW}Session Changes:${RESET}"
        for change in "${session_changes[@]}"; do
            echo "  - $change"
        done
    fi
    
    echo
    echo "Actions:"
    echo "${YELLOW}[t]${RESET} Toggle command    ${YELLOW}[r]${RESET} Rename command"
    echo "${YELLOW}[s]${RESET} Save and exit     ${YELLOW}[q]${RESET} Quit             ${YELLOW}[h]${RESET} Help"
    echo
}

# Handle menu input and actions
handle_menu_action() {
    local commands=("$@")
    local total_commands=${#commands[@]}
    
    echo -n "Enter your choice: "
    read choice
    echo
    
    case "$choice" in
        "t"|"T")
            echo -n "Enter command number to toggle (1-$total_commands): "
            read num
            if [[ "$num" =~ ^[0-9]+$ ]] && [[ $num -ge 1 ]] && [[ $num -le $total_commands ]]; then
                local command="${commands[$((num-1))]}"
                
                # Debug: show what command was selected
                echo "Debug: Selected command '$command' at index $((num-1))"
                
                if [[ -n "$command" ]]; then
                    local display_name=$(get_command_display_name "$command")
                    local current_status=$(get_command_status "$command")
                    
                    if [[ "$current_status" == "true" ]]; then
                        if disable_command "$command"; then
                            session_changes+=("Disabled: $display_name")
                            echo "${GREEN}Disabled${RESET} $display_name"
                        fi
                    else
                        if enable_command "$command"; then
                            session_changes+=("Enabled: $display_name")
                            echo "${GREEN}Enabled${RESET} $display_name"
                        fi
                    fi
                else
                    echo "${RED}Error: Empty command name at index $((num-1))${RESET}"
                fi
            else
                echo "${RED}Invalid number.${RESET} Please enter a number between 1 and $total_commands."
            fi
            echo
            echo "Press Enter to continue..."
            read
            ;;
        "r"|"R")
            echo -n "Enter command number to rename (1-$total_commands): "
            read num
            if [[ "$num" =~ ^[0-9]+$ ]] && [[ $num -ge 1 ]] && [[ $num -le $total_commands ]]; then
                local command="${commands[$((num-1))]}"
                local current_display=$(get_command_display_name "$command")
                echo "Current name: $current_display"
                echo -n "Enter new name: "
                read new_name
                
                if [[ -n "$new_name" ]] && [[ "$new_name" != "$current_display" ]]; then
                    if rename_command "$command" "$new_name"; then
                        session_changes+=("Renamed: $current_display -> $new_name")
                        echo "${GREEN}Renamed${RESET} $current_display to $new_name"
                    fi
                else
                    echo "Name unchanged."
                fi
            else
                echo "${RED}Invalid number.${RESET} Please enter a number between 1 and $total_commands."
            fi
            echo
            echo "Press Enter to continue..."
            read
            ;;
        "d"|"D")
            echo -n "Enter command number to disable (1-$total_commands): "
            read num
            if [[ "$num" =~ ^[0-9]+$ ]] && [[ $num -ge 1 ]] && [[ $num -le $total_commands ]]; then
                local command="${commands[$((num-1))]}"
                local display_name=$(get_command_display_name "$command")
                if disable_command "$command"; then
                    session_changes+=("Disabled: $display_name")
                    echo "${GREEN}Disabled${RESET} $display_name"
                fi
            else
                echo "${RED}Invalid number.${RESET} Please enter a number between 1 and $total_commands."
            fi
            echo
            echo "Press Enter to continue..."
            read
            ;;
        "s"|"S")
            echo "${GREEN}All changes saved!${RESET}"
            if [[ ${#session_changes[@]} -gt 0 ]]; then
                echo
                echo "Summary of changes:"
                for change in "${session_changes[@]}"; do
                    echo "  - $change"
                done
            else
                echo "No changes were made."
            fi
            return 2  # Signal to exit with success
            ;;
        "q"|"Q")
            if [[ ${#session_changes[@]} -gt 0 ]]; then
                echo "${YELLOW}You have unsaved changes:${RESET}"
                for change in "${session_changes[@]}"; do
                    echo "  - $change"
                done
                echo
                echo -n "Exit without saving? [y/N]: "
                read confirm
                if [[ "$confirm" =~ ^[Yy]$ ]]; then
                    echo "Exited without saving changes."
                    return 3  # Signal to exit without saving
                else
                    return 0  # Continue in menu
                fi
            else
                echo "Exited without changes."
                return 3  # Signal to exit
            fi
            ;;
        "h"|"H")
            echo "${YELLOW}Help:${RESET}"
            echo "  ${BLUE}t${RESET} - Toggle a command enabled/disabled"
            echo "  ${BLUE}r${RESET} - Rename a command"
            echo "  ${BLUE}s${RESET} - Save all changes and exit"
            echo "  ${BLUE}q${RESET} - Quit (with confirmation if there are unsaved changes)"
            echo "  ${BLUE}h${RESET} - Show this help message"
            echo
            echo "Commands are referenced by their number (1, 2, 3, etc.)"
            echo "All changes are made in a session and saved together when you choose 's'."
            echo
            echo "Press Enter to continue..."
            read
            ;;
        "")
            # Empty input, just refresh the menu
            ;;
        *)
            echo "${RED}Invalid choice.${RESET} Please enter t, r, d, s, q, or h."
            echo "Press Enter to continue..."
            read
            ;;
    esac
    
    return 0  # Continue in menu
}

# Menu-based interactive mode
interactive_mode() {
    local commands=($(scan_commands))
    
    if [[ ${#commands[@]} -eq 0 ]]; then
        print_warning "No command files found in $COMMANDS_DIR"
        return 1
    fi
    
    # Global session changes array for sharing between functions
    session_changes=()
    
    while true; do
        display_menu "${commands[@]}"
        handle_menu_action "${commands[@]}"
        local action_result=$?
        
        case $action_result in
            2) # Save and exit
                return 0
                ;;
            3) # Quit (with or without saving)
                return 1
                ;;
            *) # Continue
                continue
                ;;
        esac
    done
}

# Check for required dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v jq >/dev/null 2>&1; then
        missing_deps+=("jq")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing required dependencies:"
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        echo
        print_info "Please install missing dependencies:"
        echo "  brew install jq"
        return 1
    fi
    
    return 0
}

# Ensure required directories exist
ensure_directories() {
    # Create command library directories
    if ! mkdir -p "$COMMANDS_DIR" 2>/dev/null; then
        print_error "Cannot create commands directory: $COMMANDS_DIR"
        print_info "Please check permissions and try again"
        return 1
    fi
    
    # Create Claude commands directory
    if ! mkdir -p "$CLAUDE_COMMANDS_DIR" 2>/dev/null; then
        print_error "Cannot create Claude commands directory: $CLAUDE_COMMANDS_DIR"
        print_info "Please check permissions for ~/.claude/commands"
        return 1
    fi
    
    return 0
}

# Clean up broken symlinks
cleanup_broken_symlinks() {
    if [[ ! -d "$CLAUDE_COMMANDS_DIR" ]]; then
        return 0
    fi
    
    local broken_count=0
    
    # Find and remove broken symlinks
    while IFS= read -r -d '' symlink; do
        if [[ -L "$symlink" ]] && [[ ! -e "$symlink" ]]; then
            local basename=$(basename "$symlink" .md)
            print_warning "Removing broken symlink: $basename"
            rm "$symlink" 2>/dev/null || true
            ((broken_count++))
        fi
    done < <(find "$CLAUDE_COMMANDS_DIR" -type l -print0 2>/dev/null)
    
    if [[ $broken_count -gt 0 ]]; then
        print_info "Cleaned up $broken_count broken symlinks"
    fi
}

# Validate configuration file
validate_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        return 0
    fi
    
    # Check if JSON is valid
    if ! jq empty "$CONFIG_FILE" >/dev/null 2>&1; then
        print_warning "Configuration file contains invalid JSON"
        print_info "Backing up and reinitializing config..."
        
        # Create backup
        local backup_file="${CONFIG_FILE}.backup.$(date +%s)"
        mv "$CONFIG_FILE" "$backup_file" 2>/dev/null || true
        print_info "Backup created: $backup_file"
        
        # Reinitialize
        echo '{"commands": {}}' | jq . > "$CONFIG_FILE"
        print_success "Configuration reinitialized"
        return 1
    fi
    
    return 0
}

# Sync configuration with actual symlinks
sync_config_with_symlinks() {
    if [[ ! -d "$CLAUDE_COMMANDS_DIR" ]]; then
        return 0
    fi
    
    local config=$(cat "$CONFIG_FILE")
    local changes_made=false
    
    # Check each command in config
    while IFS= read -r command_name; do
        if [[ -n "$command_name" ]]; then
            local display_name=$(echo "$config" | jq -r ".commands[\"$command_name\"].display_name // \"$command_name\"")
            local enabled=$(echo "$config" | jq -r ".commands[\"$command_name\"].enabled // false")
            local symlink_path="$CLAUDE_COMMANDS_DIR/$display_name.md"
            
            # If command is marked enabled but symlink doesn't exist
            if [[ "$enabled" == "true" ]] && [[ ! -L "$symlink_path" ]]; then
                print_warning "Command marked enabled but symlink missing: $display_name"
                # Mark as disabled
                config=$(echo "$config" | jq --arg name "$command_name" '.commands[$name].enabled = false')
                changes_made=true
            fi
            
            # If command is marked disabled but symlink exists
            if [[ "$enabled" == "false" ]] && [[ -L "$symlink_path" ]]; then
                print_warning "Command marked disabled but symlink exists: $display_name"
                # Remove symlink
                rm "$symlink_path" 2>/dev/null || true
                changes_made=true
            fi
        fi
    done < <(echo "$config" | jq -r '.commands | keys[]' 2>/dev/null)
    
    if [[ "$changes_made" == "true" ]]; then
        save_config "$config"
        print_info "Configuration synchronized with symlinks"
    fi
}

# Handle errors gracefully
handle_error() {
    local exit_code=$1
    local error_message="$2"
    
    print_error "$error_message"
    
    case $exit_code in
        1) print_info "Check file permissions and try again" ;;
        2) print_info "Invalid configuration detected" ;;
        3) print_info "Command operation failed" ;;
        *) print_info "An error occurred during execution" ;;
    esac
    
    exit $exit_code
}

# Initialize and run with error handling
initialize_and_run() {
    # Check dependencies first
    if ! check_dependencies; then
        handle_error 1 "Missing required dependencies"
    fi
    
    # Ensure directories exist
    if ! ensure_directories; then
        handle_error 1 "Cannot create required directories"
    fi
    
    # Load and validate configuration
    load_config
    if ! validate_config; then
        print_warning "Configuration was reset due to corruption"
    fi
    
    # Clean up any issues
    cleanup_broken_symlinks
    sync_config_with_symlinks
    
    # Run main function
    main "$@"
}

# Initialize and run
initialize_and_run "$@"