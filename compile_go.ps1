# PowerShell script to compile Go version of BSOD application
param(
    [string]$Mode = "standard"
)

# Check for enhanced parameter
$Enhanced = $false
if ($Mode -eq "enhanced") {
    $Enhanced = $true
}

Write-Host "Compiling Go version of BSOD application..." -ForegroundColor Green

# Create bin directory if it doesn't exist
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
    Write-Host "Created bin directory" -ForegroundColor Yellow
} else {
    # cleanup any existing binaries
    Remove-Item -Path "./bin/BusinessAppBSOD_go.exe" -Force -ErrorAction SilentlyContinue
    if ($Enhanced) {
        Remove-Item -Path "./bin/BusinessAppBSOD_goX.exe" -Force -ErrorAction SilentlyContinue
    }
}

# Compile standard version
go build -o bin\BusinessAppBSOD_go.exe main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Go compilation successful: bin\BusinessAppBSOD_go.exe" -ForegroundColor Green
} else {
    Write-Host "✗ Go compilation failed" -ForegroundColor Red
    exit 1
}

# Compile enhanced version if requested
if ($Enhanced) {
    Write-Host "Compiling enhanced Go version..." -ForegroundColor Yellow
    go build -o bin\BusinessAppBSOD_goX.exe main_enhanced.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Enhanced Go compilation successful: bin\BusinessAppBSOD_goX.exe" -ForegroundColor Green
    } else {
        Write-Host "✗ Enhanced Go compilation failed" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`nCompilation complete!" -ForegroundColor Green

if ($Enhanced) {
    Write-Host "Available executables:" -ForegroundColor Cyan
    Write-Host "  bin\BusinessAppBSOD_go.exe - Standard Go version" -ForegroundColor White
    Write-Host "  bin\BusinessAppBSOD_goX.exe - Enhanced Go version with event logging" -ForegroundColor White
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Cyan
    Write-Host "  .\bin\BusinessAppBSOD_go.exe <access_denied | assertion_failure | 0xC0000005>" -ForegroundColor White
    Write-Host "  .\bin\BusinessAppBSOD_goX.exe <access_denied | assertion_failure | 0xC0000005>" -ForegroundColor White
} else {
    Write-Host "Run bin\BusinessAppBSOD_go.exe to test the Go version." -ForegroundColor Cyan
    Write-Host "Usage: .\bin\BusinessAppBSOD_go.exe <access_denied | assertion_failure | 0xC0000005>" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To compile both standard and enhanced versions, run:" -ForegroundColor Yellow
    Write-Host "  .\compile_go.ps1 -Mode enhanced" -ForegroundColor White
}

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
