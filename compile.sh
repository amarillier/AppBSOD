#! /bin/sh

export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc

# Create bin directory if it doesn't exist
mkdir -p bin

# cleanup any existing binaries
rm -f bin/*

go build -o bin/BusinessAppBSOD_go.exe main.go

# 32 bit Win
#export GOARCH=386
#export CC=i686-w64-mingw32-gcc


# GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -ldflags="-w -s" -o BusinessAppBSOD.exe main.go

# set CGO_ENABLED=1
# go build -ldflags="-H=windowsgui"


#go build -buildmode=c-shared -o BusinessApp.dll busappdll.go
#go build -o BusinessApp.exe main.go

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
