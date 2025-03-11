#!/bin/bash

# Configuration
REPO_OWNER="xingty"
REPO_NAME="rcode-go"
INSTALL_DIR="$HOME"
BIN_PATH="$HOME/gcode/bin"

# Text colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print status messages
print_status() {
    echo -e "${GREEN}[STATUS]${NC} $1"
}

# Function to print error messages and exit
print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Function to print warning messages
print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check for required commands
for cmd in curl jq tar; do
    if ! command -v $cmd &> /dev/null; then
        print_error "$cmd is required but not installed. Please install it and try again."
    fi
done

# Detect platform (linux, darwin, windows)
detect_platform() {
    local platform
    case "$(uname -s)" in
        Linux*)     platform="linux";;
        Darwin*)    platform="darwin";;
        CYGWIN*|MINGW*|MSYS*) platform="windows";;
        *)          platform="unknown";;
    esac
    echo "$platform"
}

# Detect architecture (amd64, arm64, etc.)
detect_architecture() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64";;
        arm64|aarch64) arch="arm64";;
        armv7l)        arch="arm";;
        i?86)          arch="386";;
        *)             arch="unknown";;
    esac
    echo "$arch"
}

# Detect platform and architecture
PLATFORM=$(detect_platform)
ARCH=$(detect_architecture)

if [ "$PLATFORM" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    print_error "Unsupported platform ($PLATFORM) or architecture ($ARCH). Cannot determine appropriate download."
fi

print_status "Detected platform: $PLATFORM, architecture: $ARCH"

print_status "Fetching the latest release information from GitHub..."

# Get the latest release info
RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
if [ $? -ne 0 ]; then
    print_error "Failed to fetch release information from GitHub. Please check your internet connection."
fi

# Check if API rate limit exceeded
if echo "$RELEASE_INFO" | grep -q "API rate limit exceeded"; then
    print_error "GitHub API rate limit exceeded. Please try again later."
fi

# Extract the tag name
TAG_NAME=$(echo "$RELEASE_INFO" | jq -r .tag_name)
if [ -z "$TAG_NAME" ] || [ "$TAG_NAME" = "null" ]; then
    print_error "Failed to get latest release tag. Please check repository name and owner."
fi

print_status "Found latest release: $TAG_NAME"

# Find the tar.gz asset URL for this platform and architecture
FILE_PATTERN="gcode-$TAG_NAME-$PLATFORM-$ARCH.tar.gz"

print_status "Looking for file matching pattern: $FILE_PATTERN"

ASSET_URL=$(echo "$RELEASE_INFO" | jq -r --arg pattern "$FILE_PATTERN" '.assets[] | select(.name == $pattern) | .browser_download_url')
if [ -z "$ASSET_URL" ] || [ "$ASSET_URL" = "null" ]; then
    print_error "No matching release file found for your platform ($PLATFORM) and architecture ($ARCH)."
fi

print_status "Found matching release file. Downloading..."

# Create a temporary directory
TMP_DIR=$(mktemp -d)
if [ $? -ne 0 ]; then
    print_error "Failed to create temporary directory."
fi

# Download the asset
TARBALL="$TMP_DIR/release.tar.gz"
curl -L -o "$TARBALL" "$ASSET_URL"
if [ $? -ne 0 ]; then
    print_error "Failed to download the release file."
fi

print_status "Download complete. Extracting to $INSTALL_DIR..."

# Extract the tarball to the install directory
tar -xzf "$TARBALL" -C "$INSTALL_DIR"
if [ $? -ne 0 ]; then
    print_error "Failed to extract the archive to $INSTALL_DIR."
fi

# Clean up
rm -rf "$TMP_DIR"

print_status "Installation completed successfully!"

# Check if the bin directory exists
if [ ! -d "$BIN_PATH" ]; then
    print_warning "The bin directory $BIN_PATH does not exist. Please check the extracted contents."
else
    print_status "Please add the following to your shell configuration file (.bashrc, .zshrc, etc.):"
    echo 'export PATH="$PATH:'$BIN_PATH'"'
    
    # Detect shell and suggest the right file
    SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
        bash)
            CONFIG_FILE="~/.bashrc"
            ;;
        zsh)
            CONFIG_FILE="~/.zshrc"
            ;;
        *)
            CONFIG_FILE="your shell configuration file"
            ;;
    esac
    
    print_status "You can do this by running:"
    echo "echo 'export PATH=\"\$PATH:$BIN_PATH\"' >> $CONFIG_FILE"
    echo "source $CONFIG_FILE"
    
    print_status "After updating your PATH, you can run the program by typing 'gcode' in your terminal."
fi

print_status "Installation process complete!"