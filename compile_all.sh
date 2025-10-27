#!/bin/bash

# Compile C and Go version of BSOD application
echo "Compiling C and Go version of BSOD application..."

# Create bin directory if it doesn't exist
mkdir -p bin

# cleanup any existing binaries
rm -f bin/BusinessAppBSOD.exe

echo "FIRST: Compiling C version of BSOD application..."
x86_64-w64-mingw32-gcc -o bin/BusinessAppBSOD.exe bsod.c -lntdll
if [ $? -eq 0 ]; then
    echo "✓ C compilation successful: bin/BusinessAppBSOD.exe\n"
else
    echo "✗ C compilation failed\n"
fi


echo "SECOND: Compiling Go version of BSOD application...\n"
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc

# cleanup any existing binaries
rm -f bin/BusinessAppBSOD_go.exe
rm -f bin/BusinessAppBSOD_goX.exe

go build -o bin/BusinessAppBSOD_go.exe main.go
if [ $? -eq 0 ]; then
    echo "✓ Go compilation successful: bin/BusinessAppBSOD_go.exe\n"
else
    echo "✗ Go compilation failed\n"
fi
# Compile enhanced version also
go build -o bin/BusinessAppBSOD_goX.exe main_enhanced.go 
if [ $? -eq 0 ]; then
    echo "✓ Enhanced Go compilation successful: bin/BusinessAppBSOD_goX.exe\n"
else
    echo "✗ Enhanced Go compilation failed"
fi

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
