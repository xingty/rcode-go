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
for cmd in curl tar; do
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

version_compare() {
    IFS='.' read -ra VER1 <<< "$1"
    IFS='.' read -ra VER2 <<< "$2"
    
    local max_length=$(( ${#VER1[@]} > ${#VER2[@]} ? ${#VER1[@]} : ${#VER2[@]} ))
    
    for (( i=0; i<max_length; i++ )); do
        local v1=${VER1[i]:-0}
        local v2=${VER2[i]:-0}
        
        if ((10#$v1 > 10#$v2)); then
            return "1"
        elif ((10#$v1 < 10#$v2)); then
            return "2"
        fi
    done
    
    return "0"
}

# Detect platform and architecture
PLATFORM=$(detect_platform)
ARCH=$(detect_architecture)

if [ "$PLATFORM" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    print_error "Unsupported platform ($PLATFORM) or architecture ($ARCH). Cannot determine appropriate download."
fi

print_status "Detected platform: $PLATFORM, architecture: $ARCH"

if command -v gcode &> /dev/null; then
    print_status "gcode is already installed. Checking for updates..."
    CURRENT_VERSION=$(gcode -v | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
    
    if [ -z "$CURRENT_VERSION" ]; then
        print_warning "Could not determine current version. Will proceed with installation."
    else
        print_status "Current version: $CURRENT_VERSION"
    fi
else
    print_status "gcode is not installed. Will proceed with installation."
    CURRENT_VERSION=""
fi

print_status "Fetching the latest version information..."

# Get the latest version from VERSION file
VERSION=$(curl -s "https://raw.githubusercontent.com/xingty/rcode-go/refs/heads/main/VERSION")
if [ $? -ne 0 ]; then
    print_error "Failed to fetch version information. Please check your internet connection."
fi

# Validate version format
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    print_error "Invalid version format: $VERSION"
fi

TAG_NAME="v$VERSION"
print_status "Found latest version: $TAG_NAME"

if [ -n "$CURRENT_VERSION" ]; then
    version_compare "$CURRENT_VERSION" "$VERSION"
    cmp=$?
    if [ $cmp -eq 0 ]; then
        print_status "gcode is already at the latest version. No update necessary."
        exit 0
    elif [ $cmp -eq 1 ]; then
        print_status "Installed gcode version ($CURRENT_VERSION) is ahead of the latest release ($VERSION)."
        exit 0
    else
        print_status "Update available: $CURRENT_VERSION -> $VERSION"
    fi
fi

FILE_PATTERN="gcode-$TAG_NAME-$PLATFORM-$ARCH.tar.gz"
print_status "Looking for file matching pattern: $FILE_PATTERN"

ASSET_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$TAG_NAME/$FILE_PATTERN"

print_status "Downloading release file from: $ASSET_URL"

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