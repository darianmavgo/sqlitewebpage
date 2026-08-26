#!/usr/bin/env bash
set -e

# ==============================================================================
# sqldoc & sqldoc-viewer Installation Script
# ==============================================================================

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

# Defaults
INSTALL_ALL=true
VIEWER_ONLY=false
CLI_ONLY=false
UNINSTALL=false
BIN_DIR=""
APP_DIR=""

# Color helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

print_help() {
    echo -e "${BOLD}sqldoc-viewer Installer${NC}"
    echo ""
    echo "Usage: ./install.sh [options]"
    echo ""
    echo "Options:"
    echo "  --viewer-only       Install only the sqldoc-viewer desktop application"
    echo "  --cli-only          Install command-line binaries only (skip macOS .app bundle)"
    echo "  --bin-dir <path>    Target directory for command-line binaries (e.g. /usr/local/bin, ~/.local/bin)"
    echo "  --app-dir <path>    Target directory for macOS .app bundle (default: /Applications or ~/Applications)"
    echo "  --uninstall         Remove installed binaries and .app bundle"
    echo "  -h, --help          Show this help message"
    echo ""
}

# Parse flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        --viewer-only)
            VIEWER_ONLY=true
            INSTALL_ALL=false
            shift
            ;;
        --cli-only)
            CLI_ONLY=true
            shift
            ;;
        --bin-dir)
            BIN_DIR="$2"
            shift 2
            ;;
        --app-dir)
            APP_DIR="$2"
            shift 2
            ;;
        --uninstall)
            UNINSTALL=true
            shift
            ;;
        -h|--help)
            print_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            print_help
            exit 1
            ;;
    esac
done

# Determine Binary Installation Directory
if [ -z "$BIN_DIR" ]; then
    if [ -w "/usr/local/bin" ]; then
        BIN_DIR="/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ]; then
        BIN_DIR="$HOME/.local/bin"
    elif [ -d "$HOME/bin" ]; then
        BIN_DIR="$HOME/bin"
    else
        BIN_DIR="$HOME/.local/bin"
    fi
fi

# Determine macOS Application Directory
if [ -z "$APP_DIR" ]; then
    if [ -w "/Applications" ]; then
        APP_DIR="/Applications"
    else
        APP_DIR="$HOME/Applications"
    fi
fi

# ==============================================================================
# UNINSTALLATION FLOW
# ==============================================================================
if [ "$UNINSTALL" = true ]; then
    echo -e "${BOLD}${BLUE}Uninstalling sqldoc suite...${NC}"
    
    # Remove binaries
    for bin in sqldoc sqldoc-serve sqldoc-viewer; do
        if [ -f "$BIN_DIR/$bin" ]; then
            echo "  Removing $BIN_DIR/$bin"
            rm -f "$BIN_DIR/$bin"
        fi
        if [ -f "/usr/local/bin/$bin" ] && [ "$BIN_DIR" != "/usr/local/bin" ]; then
            echo "  Removing /usr/local/bin/$bin"
            rm -f "/usr/local/bin/$bin" 2>/dev/null || sudo rm -f "/usr/local/bin/$bin" || true
        fi
    done

    # Remove macOS app bundle
    if [ -d "$APP_DIR/Sqldoc Viewer.app" ]; then
        echo "  Removing $APP_DIR/Sqldoc Viewer.app"
        rm -rf "$APP_DIR/Sqldoc Viewer.app"
    fi
    if [ -d "/Applications/Sqldoc Viewer.app" ] && [ "$APP_DIR" != "/Applications" ]; then
        echo "  Removing /Applications/Sqldoc Viewer.app"
        rm -rf "/Applications/Sqldoc Viewer.app" 2>/dev/null || sudo rm -rf "/Applications/Sqldoc Viewer.app" || true
    fi

    echo -e "${GREEN}✓ Uninstallation complete.${NC}"
    exit 0
fi

# ==============================================================================
# PREREQUISITES CHECK
# ==============================================================================
echo -e "${BOLD}${BLUE}Checking prerequisites...${NC}"

if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}Error: Go compiler is not installed or not in PATH.${NC}"
    echo "Please install Go (https://golang.org/dl/) and retry."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "  Found $GO_VERSION"

# Ensure target bin directory exists
mkdir -p "$BIN_DIR"

# ==============================================================================
# BUILD BINARIES
# ==============================================================================
echo -e "\n${BOLD}${BLUE}Building binaries...${NC}"
mkdir -p bin

echo "  Building sqldoc-viewer (native WebKit GUI)..."
CGO_ENABLED=1 go build -o bin/sqldoc-viewer ./cmd/sqldoc-viewer

if [ "$VIEWER_ONLY" = false ]; then
    echo "  Building sqldoc (CLI renderer & inspector)..."
    go build -o bin/sqldoc ./cmd/sqldoc

    echo "  Building sqldoc-serve (local HTTP server)..."
    go build -o bin/sqldoc-serve ./cmd/sqldoc-serve
fi

# ==============================================================================
# INSTALL CLI BINARIES
# ==============================================================================
echo -e "\n${BOLD}${BLUE}Installing CLI binaries to ${BIN_DIR}...${NC}"

install_bin() {
    local src="$1"
    local dest="$2"
    if [ -w "$(dirname "$dest")" ]; then
        cp -f "$src" "$dest"
        chmod +x "$dest"
    else
        echo "  Requesting sudo permissions to install to $dest..."
        sudo cp -f "$src" "$dest"
        sudo chmod +x "$dest"
    fi
    echo -e "  ${GREEN}✓ Installed $(basename "$dest")${NC}"
}

install_bin "bin/sqldoc-viewer" "$BIN_DIR/sqldoc-viewer"

if [ "$VIEWER_ONLY" = false ]; then
    install_bin "bin/sqldoc" "$BIN_DIR/sqldoc"
    install_bin "bin/sqldoc-serve" "$BIN_DIR/sqldoc-serve"
fi

# ==============================================================================
# CREATE MACOS APPLICATION BUNDLE (if on macOS and not --cli-only)
# ==============================================================================
if [[ "$(uname)" == "Darwin" && "$CLI_ONLY" = false ]]; then
    echo -e "\n${BOLD}${BLUE}Creating macOS Application Bundle in ${APP_DIR}...${NC}"
    mkdir -p "$APP_DIR"

    APP_BUNDLE="$APP_DIR/Sqldoc Viewer.app"
    CONTENTS="$APP_BUNDLE/Contents"
    MACOS_DIR="$CONTENTS/MacOS"
    RESOURCES_DIR="$CONTENTS/Resources"

    rm -rf "$APP_BUNDLE"
    mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"

    # Copy binary into app bundle
    cp -f "bin/sqldoc-viewer" "$MACOS_DIR/sqldoc-viewer"
    chmod +x "$MACOS_DIR/sqldoc-viewer"

    # Generate Info.plist
    cat > "$CONTENTS/Info.plist" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>sqldoc-viewer</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.sqldoc.viewer</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>Sqldoc Viewer</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>CFBundleDocumentTypes</key>
    <array>
        <dict>
            <key>CFBundleTypeName</key>
            <string>SQLite Database</string>
            <key>CFBundleTypeRole</key>
            <string>Viewer</string>
            <key>LSHandlerRank</key>
            <string>Alternate</string>
            <key>CFBundleTypeExtensions</key>
            <array>
                <string>db</string>
                <string>sqlite</string>
                <string>sqlite3</string>
                <string>sqldoc</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
EOF

    # Generate High-Resolution macOS Icon (.icns)
    if command -v sips >/dev/null 2>&1 && command -v iconutil >/dev/null 2>&1; then
        ICONSET_DIR=$(mktemp -d)/AppIcon.iconset
        mkdir -p "$ICONSET_DIR"

        # Create a base 1024x1024 icon PNG using python & core graphics / PIL if available, or generate a clean SVG
        python3 - << 'PYEOF'
import os
import struct
import zlib

def create_icon(filename):
    # Minimal pure Python PNG generator for a 512x512 gradient icon with database icon
    # Or create an SVG and rasterize via QuickLook/WebKit
    svg_content = '''<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024" viewBox="0 0 1024 1024">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#4f46e5" />
      <stop offset="100%" stop-color="#2563eb" />
    </linearGradient>
    <filter id="shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="24" stdDeviation="32" flood-opacity="0.3"/>
    </filter>
  </defs>
  <rect x="64" y="64" width="896" height="896" rx="200" fill="url(#bg)" filter="url(#shadow)" />
  <g fill="none" stroke="#ffffff" stroke-width="48" stroke-linecap="round" stroke-linejoin="round" transform="translate(192, 192) scale(0.625)">
    <!-- Top Cylinder -->
    <ellipse cx="512" cy="220" rx="360" ry="140" fill="#ffffff" fill-opacity="0.2"/>
    <ellipse cx="512" cy="220" rx="360" ry="140"/>
    
    <!-- Middle Cylinder Body -->
    <path d="M 152,220 V 480 C 152,620 872,620 872,480 V 220" fill="#ffffff" fill-opacity="0.15"/>
    <path d="M 152,480 C 152,620 872,620 872,480"/>
    
    <!-- Bottom Cylinder Body -->
    <path d="M 152,480 V 740 C 152,880 872,880 872,740 V 480" fill="#ffffff" fill-opacity="0.1"/>
    <path d="M 152,740 C 152,880 872,880 872,740"/>
  </g>
</svg>'''
    with open("/tmp/sqldoc_icon.svg", "w") as f:
        f.write(svg_content)
PYEOF

        if [ -f "/tmp/sqldoc_icon.svg" ]; then
            qlmanage -t -s 1024 -o /tmp /tmp/sqldoc_icon.svg >/dev/null 2>&1 || true
            BASE_PNG="/tmp/sqldoc_icon.svg.png"
            if [ ! -f "$BASE_PNG" ]; then
                BASE_PNG="/tmp/sqldoc_icon.png"
                # Fallback: create base PNG with sips if needed
            fi

            if [ -f "$BASE_PNG" ]; then
                sips -z 16 16     "$BASE_PNG" --out "$ICONSET_DIR/icon_16x16.png" >/dev/null 2>&1
                sips -z 32 32     "$BASE_PNG" --out "$ICONSET_DIR/icon_16x16@2x.png" >/dev/null 2>&1
                sips -z 32 32     "$BASE_PNG" --out "$ICONSET_DIR/icon_32x32.png" >/dev/null 2>&1
                sips -z 64 64     "$BASE_PNG" --out "$ICONSET_DIR/icon_32x32@2x.png" >/dev/null 2>&1
                sips -z 128 128   "$BASE_PNG" --out "$ICONSET_DIR/icon_128x128.png" >/dev/null 2>&1
                sips -z 256 256   "$BASE_PNG" --out "$ICONSET_DIR/icon_128x128@2x.png" >/dev/null 2>&1
                sips -z 256 256   "$BASE_PNG" --out "$ICONSET_DIR/icon_256x256.png" >/dev/null 2>&1
                sips -z 512 512   "$BASE_PNG" --out "$ICONSET_DIR/icon_256x256@2x.png" >/dev/null 2>&1
                sips -z 512 512   "$BASE_PNG" --out "$ICONSET_DIR/icon_512x512.png" >/dev/null 2>&1
                sips -z 1024 1024 "$BASE_PNG" --out "$ICONSET_DIR/icon_512x512@2x.png" >/dev/null 2>&1

                iconutil -c icns "$ICONSET_DIR" -o "$RESOURCES_DIR/AppIcon.icns" >/dev/null 2>&1 || true
            fi
            rm -rf "$ICONSET_DIR" /tmp/sqldoc_icon.svg /tmp/sqldoc_icon.svg.png /tmp/sqldoc_icon.png 2>/dev/null || true
        fi
    fi

    echo -e "  ${GREEN}✓ Created $APP_BUNDLE${NC}"
fi

# ==============================================================================
# SUMMARY & VERIFICATION
# ==============================================================================
echo -e "\n${BOLD}${GREEN}Installation Successful!${NC}"
echo "=================================================="
echo -e "Binary installed to:  ${BOLD}$BIN_DIR/sqldoc-viewer${NC}"
if [ "$VIEWER_ONLY" = false ]; then
    echo -e "Companion CLI tools:  ${BOLD}$BIN_DIR/sqldoc${NC}"
    echo -e "                      ${BOLD}$BIN_DIR/sqldoc-serve${NC}"
fi
if [[ "$(uname)" == "Darwin" && "$CLI_ONLY" = false ]]; then
    echo -e "macOS Application:    ${BOLD}$APP_DIR/Sqldoc Viewer.app${NC}"
fi
echo "=================================================="

# Check if BIN_DIR is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "\n${RED}Note:${NC} $BIN_DIR is not in your current PATH."
    echo "Add the following line to your ~/.zshrc or ~/.bashrc:"
    echo -e "  ${BOLD}export PATH=\"\$PATH:$BIN_DIR\"${NC}"
fi

echo -e "\n${BOLD}Quick Test:${NC}"
echo "  sqldoc-viewer examples/demo.db"
if [[ "$(uname)" == "Darwin" && "$CLI_ONLY" = false ]]; then
    echo "  open -a \"Sqldoc Viewer\" examples/demo.db"
fi
echo ""
