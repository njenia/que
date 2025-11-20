#!/bin/bash
# Que installer script
# Downloads and installs the latest release of Que

set -e

REPO="njenia/que"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="que"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$OS" == "linux" ]]; then
  OS="linux"
elif [[ "$OS" == "darwin" ]]; then
  OS="darwin"
else
  echo -e "${RED}Error: Unsupported OS: $OS${NC}"
  exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
    exit 1
    ;;
esac

# Determine file extension
if [[ "$OS" == "darwin" || "$OS" == "linux" ]]; then
  EXT="tar.gz"
else
  EXT="zip"
fi

# Use the /latest/download/ endpoint which automatically redirects to the latest release
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/que-${OS}-${ARCH}.${EXT}"
VERSION="latest"

if [[ -n "$VERSION" ]]; then
  echo -e "${GREEN}Installing Que ${VERSION} for ${OS}/${ARCH}...${NC}"
else
  echo -e "${GREEN}Installing Que (latest) for ${OS}/${ARCH}...${NC}"
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
echo -e "${YELLOW}Downloading from ${DOWNLOAD_URL}...${NC}"
cd "$TMP_DIR"

if command -v curl >/dev/null 2>&1; then
  HTTP_CODE=$(curl -sL -w "%{http_code}" -o "que.${EXT}" "$DOWNLOAD_URL" || echo "000")
  if [[ "$HTTP_CODE" != "200" ]]; then
    echo -e "${RED}Error: Failed to download (HTTP $HTTP_CODE)${NC}"
    if [[ -f "que.${EXT}" ]]; then
      echo -e "${YELLOW}Response:$(cat que.${EXT})${NC}"
      rm -f "que.${EXT}"
    fi
    exit 1
  fi
elif command -v wget >/dev/null 2>&1; then
  if ! wget -q -O "que.${EXT}" "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Failed to download${NC}"
    exit 1
  fi
else
  echo -e "${RED}Error: Neither curl nor wget is installed${NC}"
  exit 1
fi

# Verify downloaded file size
FILE_SIZE=$(stat -f%z "que.${EXT}" 2>/dev/null || stat -c%s "que.${EXT}" 2>/dev/null || echo "0")
if [[ "$FILE_SIZE" -lt 1000 ]]; then
  echo -e "${RED}Error: Downloaded file is too small (${FILE_SIZE} bytes), likely an error page${NC}"
  echo -e "${YELLOW}Response:$(head -c 200 que.${EXT})${NC}"
  rm -f "que.${EXT}"
  exit 1
fi

# Extract
if [[ "$EXT" == "tar.gz" ]]; then
  tar -xzf "que.${EXT}"
else
  unzip -q "que.${EXT}"
fi

# Check if binary exists
if [[ ! -f "$BINARY_NAME" ]]; then
  echo -e "${RED}Error: Binary not found after extraction${NC}"
  exit 1
fi

# Make binary executable
chmod +x "$BINARY_NAME"

# Install to target directory
if [[ ! -w "$INSTALL_DIR" ]]; then
  echo -e "${YELLOW}Note: ${INSTALL_DIR} requires sudo permissions${NC}"
  sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
  mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
  INSTALLED_VERSION=$("$BINARY_NAME" --version 2>&1 | head -n 1 || echo "unknown")
  echo -e "${GREEN}✓ Que installed successfully!${NC}"
  echo -e "${GREEN}  Version: ${INSTALLED_VERSION}${NC}"
  echo -e "${GREEN}  Location: $(which $BINARY_NAME)${NC}"
else
  echo -e "${YELLOW}Warning: Que installed but not found in PATH${NC}"
  echo -e "${YELLOW}  Installed to: ${INSTALL_DIR}/${BINARY_NAME}${NC}"
  echo -e "${YELLOW}  Make sure ${INSTALL_DIR} is in your PATH${NC}"
fi

