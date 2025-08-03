#!/bin/bash
# Change to the directory containing this script
cd "$(dirname "$0")"

echo "Testing Go build for Claude Command Manager"
echo "Current directory: $(pwd)"
echo

echo "Checking Go installation..."
go version
echo

echo "Verifying go.mod..."
cat go.mod | head -5
echo

echo "Building application..."
go build -o claude_command_manager cmd/main.go
BUILD_STATUS=$?

if [ $BUILD_STATUS -eq 0 ]; then
    echo "✅ Build successful!"
    echo
    
    echo "Testing help command..."
    ./claude_command_manager help
    echo
    
    echo "Checking if binary was created..."
    ls -la claude_command_manager
    
    echo "✅ All tests passed!"
else
    echo "❌ Build failed with status $BUILD_STATUS"
fi

echo
echo "Cleaning up..."
rm -f claude_command_manager