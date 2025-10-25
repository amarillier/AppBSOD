#! /bin/sh

# Check for enhanced parameter
ENHANCED=false
if [ "$1" = "enhanced" ]; then
    ENHANCED=true
fi

export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc

# Create bin directory if it doesn't exist
mkdir -p bin

# cleanup any existing binaries
rm -f bin/BusinessAppBSOD_go.exe
if [ "$ENHANCED" = true ]; then
    rm -f bin/BusinessAppBSOD_goX.exe
fi

echo "Compiling Go version of BSOD application..."
go build -o bin/BusinessAppBSOD_go.exe main.go

if [ $? -eq 0 ]; then
    echo "✓ Go compilation successful: bin/BusinessAppBSOD_go.exe"
else
    echo "✗ Go compilation failed"
    exit 1
fi

# Compile enhanced version if requested
if [ "$ENHANCED" = true ]; then
    echo "Compiling enhanced Go version..."
    go build -o bin/BusinessAppBSOD_goX.exe main_enhanced.go
    
    if [ $? -eq 0 ]; then
        echo "✓ Enhanced Go compilation successful: bin/BusinessAppBSOD_goX.exe"
    else
        echo "✗ Enhanced Go compilation failed"
        exit 1
    fi
fi

# 32 bit Win
#export GOARCH=386
#export CC=i686-w64-mingw32-gcc


# GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -ldflags="-w -s" -o BusinessAppBSOD.exe main.go

# set CGO_ENABLED=1
# go build -ldflags="-H=windowsgui"


echo ""
echo "Compilation complete!"
if [ "$ENHANCED" = true ]; then
    echo "Available executables:"
    echo "  bin/BusinessAppBSOD_go.exe - Standard Go version"
    echo "  bin/BusinessAppBSOD_goX.exe - Enhanced Go version with event logging"
    echo ""
    echo "Usage:"
    echo "  ./bin/BusinessAppBSOD_go.exe <access_denied | assertion_failure | 0xC0000005>"
    echo "  ./bin/BusinessAppBSOD_goX.exe <access_denied | assertion_failure | 0xC0000005>"
else
    echo "Run bin/BusinessAppBSOD_go.exe to test the Go version."
    echo "Usage: ./bin/BusinessAppBSOD_go.exe <access_denied | assertion_failure | 0xC0000005>"
    echo ""
    echo "To compile both standard and enhanced versions, run:"
    echo "  ./compile_go.sh enhanced"
fi

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
