#!/bin/bash

# Compile C version of BSOD application
echo "Compiling C version of BSOD application..."

# Create bin directory if it doesn't exist
mkdir -p bin

# cleanup any existing binaries
rm -f bin/BusinessAppBSOD.exe

# 64-bit Windows
x86_64-w64-mingw32-gcc -o bin/BusinessAppBSOD.exe bsod.c -lntdll

# 32-bit Windows (commented out)
# i686-w64-mingw32-gcc -o BusinessAppBSOD_32.exe bsod.c -lntdll

echo "Compilation complete. Run bin/BusinessAppBSOD.exe to test."

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
