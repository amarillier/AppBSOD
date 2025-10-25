# PowerShell script to compile C version of BSOD application using MinGW GCC

Write-Host "Compiling C version of BSOD application using MinGW GCC..." -ForegroundColor Green

# Check if MinGW GCC is available
try {
    $gccVersion = & x86_64-w64-mingw32-gcc --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Found MinGW GCC:" -ForegroundColor Yellow
        Write-Host $gccVersion[0] -ForegroundColor Cyan
    } else {
        throw "MinGW GCC not found"
    }
} catch {
    Write-Host "Error: MinGW GCC (x86_64-w64-mingw32-gcc) not found in PATH" -ForegroundColor Red
    Write-Host "Please install MinGW-w64 or add it to your PATH" -ForegroundColor Yellow
    Write-Host "You can install it via:" -ForegroundColor Yellow
    Write-Host "  - MSYS2: pacman -S mingw-w64-x86_64-gcc" -ForegroundColor Cyan
    Write-Host "  - Chocolatey: choco install mingw" -ForegroundColor Cyan
    Write-Host "  - Or download from: https://www.mingw-w64.org/" -ForegroundColor Cyan
    exit 1
}

# Create bin directory if it doesn't exist
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
    Write-Host "Created bin directory" -ForegroundColor Yellow
} else {
    # cleanup any existing binaries
    Remove-Item -Path "./bin/BusinessAppBSOD.exe" -Force
}

# Compile 64-bit Windows executable
Write-Host "`nCompiling 64-bit Windows executable..." -ForegroundColor Yellow
& x86_64-w64-mingw32-gcc -o bin\BusinessAppBSOD.exe bsod.c -lntdll

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 64-bit compilation successful: BusinessAppBSOD.exe" -ForegroundColor Green
} else {
    Write-Host "✗ 64-bit compilation failed" -ForegroundColor Red
    exit 1
}

# Optional: Compile 32-bit version (commented out by default)
# Uncomment the following lines if you want to compile a 32-bit version
<#
Write-Host "`nCompiling 32-bit Windows executable..." -ForegroundColor Yellow
& i686-w64-mingw32-gcc -o BusinessAppBSOD_C_32.exe bsod.c -lntdll

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 32-bit compilation successful: BusinessAppBSOD_C_32.exe" -ForegroundColor Green
} else {
    Write-Host "✗ 32-bit compilation failed" -ForegroundColor Red
}
#>

Write-Host "`nCompilation complete!" -ForegroundColor Green
Write-Host "Run BusinessAppBSOD.exe to test the C version." -ForegroundColor Cyan
Write-Host "Usage: .\bin\BusinessAppBSOD.exe <access_denied | assertion_failure | 0xC0000005>" -ForegroundColor Cyan

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
