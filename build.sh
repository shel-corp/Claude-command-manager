#!/usr/bin/env bash

set -e

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

# Check if Go is installed
check_go() {
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed or not in PATH"
        print_info "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Found Go version: $go_version"
}

# Build for current platform
build_current() {
    print_info "Building for current platform..."
    go build -o ccm cmd/main.go
    print_success "Built: ccm"
    
    # Make executable
    chmod +x ccm
}

# Build for multiple platforms
build_all() {
    print_info "Building for multiple platforms..."
    
    local platforms=(
        "darwin/amd64"
        "darwin/arm64"  
        "linux/amd64"
        "linux/arm64"
        "windows/amd64"
    )
    
    mkdir -p dist
    
    for platform in "${platforms[@]}"; do
        local os=${platform%/*}
        local arch=${platform#*/}
        local output="dist/ccm-$os-$arch"
        
        if [[ "$os" == "windows" ]]; then
            output="$output.exe"
        fi
        
        print_info "Building for $os/$arch..."
        GOOS=$os GOARCH=$arch go build -o "$output" cmd/main.go
        
        if [[ $? -eq 0 ]]; then
            print_success "Built: $output"
        else
            print_error "Failed to build for $os/$arch"
        fi
    done
}

# Clean build artifacts
clean() {
    print_info "Cleaning build artifacts..."
    rm -f ccm
    rm -rf dist/
    print_success "Cleaned build artifacts"
}

# Run tests
test() {
    print_info "Running tests..."
    go test ./...
    if [[ $? -eq 0 ]]; then
        print_success "All tests passed"
    else
        print_error "Some tests failed"
        exit 1
    fi
}

# Show usage
usage() {
    echo "Claude Command Manager Build Script"
    echo
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  build        Build for current platform (default)"
    echo "  all          Build for all platforms"
    echo "  clean        Clean build artifacts"
    echo "  test         Run tests"
    echo "  help         Show this help message"
}

# Main function
main() {
    check_go
    
    case "${1:-build}" in
        "build")
            build_current
            ;;
        "all")
            build_all
            ;;
        "clean")
            clean
            ;;
        "test")
            test
            ;;
        "help"|"-h"|"--help")
            usage
            ;;
        *)
            print_error "Unknown command: $1"
            usage
            exit 1
            ;;
    esac
}

main "$@"