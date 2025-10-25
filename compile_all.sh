#!/bin/bash

# Compile C and Go version of BSOD application
echo "Compiling C and Go version of BSOD application..."

# Create bin directory if it doesn't exist
mkdir -p bin

# cleanup any existing binaries
rm -f bin/BusinessAppBSOD.exe

echo "FIRST: Compiling C version of BSOD application..."
x86_64-w64-mingw32-gcc -o bin/BusinessAppBSOD.exe bsod.c -lntdll

echo "Compilation complete. Run bin/BusinessAppBSOD.exe to test."

echo "SECOND: Compiling Go version of BSOD application..."
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc

# cleanup any existing binaries
rm -f bin/BusinessAppBSOD_go.exe

go build -o bin/BusinessAppBSOD_go.exe main.go

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
