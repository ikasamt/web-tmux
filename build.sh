#!/bin/bash

set -e

VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building Web Terminal v${VERSION} (${GIT_COMMIT})"
echo "=============================================="

# Build Angular frontend
echo "📦 Building Angular frontend..."
cd frontend
npm run build
cd ..

# Copy frontend dist to backend for embedding
echo "📁 Copying frontend assets to backend..."
rm -rf backend/static
cp -r frontend/dist backend/static

# Create bin directory if it doesn't exist
mkdir -p bin

# Build information
LDFLAGS="-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

# Build for multiple platforms
echo "🔨 Cross-compiling binaries..."

# macOS (Apple Silicon)
echo "  → macOS ARM64..."
cd backend
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ../bin/web-terminal-darwin-arm64 main.go

# macOS (Intel)
echo "  → macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ../bin/web-terminal-darwin-amd64 main.go

# Linux (ARM64)
echo "  → Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ../bin/web-terminal-linux-arm64 main.go

# Linux (AMD64)
echo "  → Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ../bin/web-terminal-linux-amd64 main.go

# Windows (AMD64)
echo "  → Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ../bin/web-terminal-windows-amd64.exe main.go

cd ..

# Create platform-specific symlinks/copies for convenience
echo "🔗 Creating platform shortcuts..."

case "$(uname -s)" in
    Darwin)
        if [[ "$(uname -m)" == "arm64" ]]; then
            ln -sf web-terminal-darwin-arm64 bin/web-terminal
        else
            ln -sf web-terminal-darwin-amd64 bin/web-terminal
        fi
        ;;
    Linux)
        if [[ "$(uname -m)" == "aarch64" ]]; then
            ln -sf web-terminal-linux-arm64 bin/web-terminal
        else
            ln -sf web-terminal-linux-amd64 bin/web-terminal
        fi
        ;;
    MINGW*|MSYS*|CYGWIN*)
        cp bin/web-terminal-windows-amd64.exe bin/web-terminal.exe
        ;;
esac

echo ""
echo "✅ Build complete!"
echo ""
echo "📊 Built binaries:"
ls -la bin/web-terminal-*
echo ""
echo "🚀 To run on your platform:"
echo "   ./bin/web-terminal"
echo ""
echo "🌐 Access the terminal at: http://localhost:8080"
echo ""
echo "📋 Available binaries:"
echo "   macOS (ARM64):    bin/web-terminal-darwin-arm64"
echo "   macOS (Intel):    bin/web-terminal-darwin-amd64" 
echo "   Linux (ARM64):    bin/web-terminal-linux-arm64"
echo "   Linux (x86_64):   bin/web-terminal-linux-amd64"
echo "   Windows (x86_64): bin/web-terminal-windows-amd64.exe"

# Clean up (keep static directory for development, but remove in CI)
if [ "${CI}" = "true" ]; then
    echo "🧹 Cleaning up static files..."
    rm -rf backend/static
fi