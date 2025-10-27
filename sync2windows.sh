#!/bin/bash
# Sync AppBSOD to Windows via mounted share and auto-compile C version

set -e  # Exit on error

# Check if this is a sync-back operation
if [ "$1" = "sync-back" ]; then
    SYNC_BACK_ONLY=true
else
    SYNC_BACK_ONLY=false
fi

# mount the Windows share if not already mounted
if ! mount | grep "AllanMWin"
then
    echo "Not mounted, mounting"
    mount -t smbfs //allanm@allanm.marillier.local/AllanMWin ~/AllanMWin
fi

# Configuration - ADJUST THESE PATHS
WINDOWS_SHARE="$HOME/AllanMWin/Allan/Source/go/AppBSOD"  # Mounted Windows share path
WINDOWS_HOST="192.168.1.9"  # Your Windows machine IP (for PowerShell remoting)
WINDOWS_USER="allan"   # Windows username

# Application names for BSOD project
APP_NAME="BusinessAppBSOD_go"
C_APP_NAME="BusinessAppBSOD"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

if [ "$SYNC_BACK_ONLY" = false ]; then
    echo -e "${BLUE}=== Syncing AppBSOD to Windows ===${NC}"

    # Check if Windows share is mounted
    if [ ! -d "$WINDOWS_SHARE" ]; then
        echo "ERROR: Windows share not mounted at: $WINDOWS_SHARE"
        echo "Please mount your Windows share first, or update WINDOWS_SHARE path in this script"
        exit 1
    fi

    # Create directory if it doesn't exist
    mkdir -p "$WINDOWS_SHARE"

    # Clean up any existing Windows binaries from Windows share for fresh compilation
    echo "Cleaning any existing Windows binaries from Windows share..."
    binDir="$WINDOWS_SHARE/bin"
    if [ -d "$binDir" ]; then
        echo "Removing all files from Windows bin directory..."
        rm -f "$binDir"/* 2>/dev/null || true
    fi
    
    # Sync source files to Windows share (excluding bin directory entirely)
    echo "Copying source files to Windows share..."
    rsync -av \
      --exclude='bin/' \
      --exclude='.git/' \
      --exclude='.DS_Store' \
      --exclude='test/' \
      . "$WINDOWS_SHARE/"

    echo -e "${GREEN}✓ Files synced to: $WINDOWS_SHARE${NC}"
    echo ""

    echo -e "${BLUE}=== Next Steps ===${NC}"
    echo "Files have been synced to Windows. Now you need to compile them:"
    echo ""
    echo "1. Go to your Windows machine"
    echo "2. Open PowerShell"
    echo "3. Navigate to: C:\\Allan\\Source\\go\\AppBSOD"
    echo ""
    echo "For Go version:"
    echo "  .\\compile_go.ps1"
    echo ""
    echo "For C version (recommended for Windows Event Logs):"
    echo "  .\\compile_c.ps1"
    echo ""
    echo "4. After compiling, come back to Mac and run:"
    echo "   ./sync2windows.sh sync-back"
    echo "   to retrieve the compiled binaries"
else
    echo -e "${BLUE}=== Syncing back compiled binaries from Windows ===${NC}"
fi

if [ "$SYNC_BACK_ONLY" = true ]; then
    # Sync-back mode: only sync binaries, don't touch source files
    echo -e "${BLUE}=== Syncing back compiled binaries from Windows ===${NC}"
    
    echo "Checking what's in Windows share bin directory..."
    ls -la "$WINDOWS_SHARE/bin/" 2>/dev/null || echo "No bin directory found in Windows share"
    
    # Ensure local bin directory exists
    mkdir -p "./bin"
    
    # Sync only the bin directory from Windows
    echo "Syncing Windows binaries..."
    rsync -av --update --exclude="*.h" --exclude="*.res" "$WINDOWS_SHARE/bin/" "./bin/" 2>/dev/null || echo "No bin directory to sync"
    
    # Clean up any non-Windows binaries that might have been synced
    # rm -f ./bin/*-linux-* ./bin/*-macos-* ./bin/*.dylib ./bin/*.so 2>/dev/null || true
    
    echo "Checking what was synced back to local bin directory..."
    ls -la "./bin/" 2>/dev/null || echo "No local bin directory found"
    
    # Fallback: explicitly copy Windows binaries if rsync didn't work
    echo "Attempting fallback copy of Windows binaries..."
    if [ -f "$WINDOWS_SHARE/bin/${APP_NAME}.exe" ]; then
        echo "Copying ${APP_NAME}.exe..."
        cp "$WINDOWS_SHARE/bin/${APP_NAME}.exe" "./bin/" 2>/dev/null || echo "Failed to copy ${APP_NAME}.exe"
    fi
    
    if [ -f "$WINDOWS_SHARE/bin/${C_APP_NAME}.exe" ]; then
        echo "Copying ${C_APP_NAME}.exe..."
        cp "$WINDOWS_SHARE/bin/${C_APP_NAME}.exe" "./bin/" 2>/dev/null || echo "Failed to copy ${C_APP_NAME}.exe"
    fi
    
    echo "Final check of local bin directory..."
    ls -la "./bin/" 2>/dev/null || echo "No local bin directory found"
    
    echo -e "${GREEN}✓ Binaries synced back from Windows${NC}"
else
    # Normal mode: sync source files and binaries
    echo -e "${BLUE}=== Syncing changes back from Windows ===${NC}"
    
    # Sync back source files (excluding bin directory to preserve local binaries)
    echo "Syncing back source files..."
    rsync -av --update \
      --exclude='bin/' \
      --exclude='test/' \
      --exclude='.git/' \
      --exclude='.DS_Store' \
      "$WINDOWS_SHARE/" .

    echo "Syncing back Windows binaries only..."
    echo "Checking what's in Windows share bin directory..."
    ls -la "$WINDOWS_SHARE/bin/" 2>/dev/null || echo "No bin directory found in Windows share"

    # Ensure local bin directory exists
    mkdir -p "./bin"

    # Only sync Windows binaries (.exe files)
    if [ -f "$WINDOWS_SHARE/bin/${APP_NAME}.exe" ]; then
        echo "Copying ${APP_NAME}.exe..."
        cp "$WINDOWS_SHARE/bin/${APP_NAME}.exe" "./bin/" 2>/dev/null || echo "Failed to copy ${APP_NAME}.exe"
    fi
    
    if [ -f "$WINDOWS_SHARE/bin/${C_APP_NAME}.exe" ]; then
        echo "Copying ${C_APP_NAME}.exe..."
        cp "$WINDOWS_SHARE/bin/${C_APP_NAME}.exe" "./bin/" 2>/dev/null || echo "Failed to copy ${C_APP_NAME}.exe"
    fi
    
    # Check for any other Windows binaries that might have been created
    for file in "$WINDOWS_SHARE/bin/"*.exe; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            echo "Copying $filename..."
            cp "$file" "./bin/" 2>/dev/null || echo "Failed to copy $filename"
        fi
    done
    
    echo "Final check of local bin directory..."
    ls -la "./bin/" 2>/dev/null || echo "No local bin directory found"
    
    echo -e "${GREEN}✓ Files synced back from Windows${NC}"
fi
echo ""

# Check for compiled binaries and show info
echo -e "${BLUE}=== Checking for compiled binaries ===${NC}"

# Check for Go BSOD executable
if [ -f "./bin/${APP_NAME}.exe" ]; then
    echo -e "${GREEN}✓ Go AppBSOD executable found: ./bin/${APP_NAME}.exe${NC}"
    ls -lh ./bin/${APP_NAME}.exe
    echo ""
    echo "Go-based BSOD testing:"
    echo "  .\\bin\\${APP_NAME}.exe access_denied"
    echo "  .\\bin\\${APP_NAME}.exe assertion_failure"
    echo "  .\\bin\\${APP_NAME}.exe 0xC0000005"
    echo "  .\\bin\\${APP_NAME}.exe 0xDEADBEEF"
    echo ""
else
    echo -e "${YELLOW}⚠ Go AppBSOD executable not found${NC}"
    echo "This means Go compilation failed or wasn't completed on Windows."
fi

# Check for C BSOD executable
if [ -f "./bin/${C_APP_NAME}.exe" ]; then
    echo -e "${GREEN}✓ C AppBSOD executable found: ./bin/${C_APP_NAME}.exe${NC}"
    ls -lh ./bin/${C_APP_NAME}.exe
    echo ""
    echo "C-based BSOD testing (recommended for Windows Event Logs):"
    echo "  .\\bin\\${C_APP_NAME}.exe access_denied"
    echo "  .\\bin\\${C_APP_NAME}.exe assertion_failure"
    echo "  .\\bin\\${C_APP_NAME}.exe 0xC0000005"
    echo "  .\\bin\\${C_APP_NAME}.exe 0xDEADBEEF"
    echo ""
else
    echo -e "${YELLOW}⚠ C AppBSOD executable not found${NC}"
    echo "This means C compilation failed or wasn't completed on Windows."
    echo ""
    echo "To fix this:"
    echo "1. Go to Windows machine"
    echo "2. Open PowerShell"
    echo "3. cd to C:\\Allan\\Source\\go\\AppBSOD"
    echo "4. Run: .\\compile_c.ps1"
    echo "5. Run this sync script again"
fi

echo ""
echo -e "${GREEN}=== Sync Complete ===${NC}"
echo ""
echo "Files are at: $WINDOWS_SHARE"
echo ""
echo "Next steps:"
echo "1. Test BSOD: .\\bin\\${C_APP_NAME}.exe access_denied"
echo "2. Check Event Logs: .\\check_system_config.ps1"
echo "3. Check minidump files in C:\\Windows\\Minidump"

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
