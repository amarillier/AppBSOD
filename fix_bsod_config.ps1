# PowerShell script to fix Windows BSOD logging configuration
Write-Host "Fixing Windows BSOD logging configuration..." -ForegroundColor Green

# Check if running as administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "ERROR: This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator'" -ForegroundColor Yellow
    exit 1
}

Write-Host "Setting up proper BSOD logging configuration..." -ForegroundColor Yellow

# Set WriteEventToLog to 1 (enable event logging)
Write-Host "Enabling event logging..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -Name "WriteEventToLog" -Value 1 -Type DWord

# Set DumpType to 1 (Small memory dump - 256KB)
# DumpType values: 0=None, 1=Small, 2=Kernel, 3=Complete, 4=Automatic, 5=Active, 6=Bitmap, 7=Automatic with kernel
Write-Host "Setting dump type to Small memory dump (256KB)..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -Name "DumpType" -Value 1 -Type DWord

# Ensure minidump directory exists
$minidumpDir = "C:\Windows\Minidump"
if (-not (Test-Path $minidumpDir)) {
    Write-Host "Creating minidump directory..." -ForegroundColor Cyan
    New-Item -Path $minidumpDir -ItemType Directory -Force | Out-Null
}

# Set minidump directory path
Write-Host "Setting minidump directory..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -Name "MinidumpDir" -Value $minidumpDir -Type ExpandString

# Set dump file location
Write-Host "Setting dump file location..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -Name "DumpFile" -Value "C:\Windows\MEMORY.DMP" -Type ExpandString

# Enable automatic restart after BSOD (optional)
Write-Host "Enabling automatic restart after BSOD..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -Name "AutoReboot" -Value 1 -Type DWord

# Set crash dump behavior
Write-Host "Setting crash dump behavior..." -ForegroundColor Cyan
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -Name "CrashDumpEnabled" -Value 1 -Type DWord

Write-Host "`nVerifying configuration..." -ForegroundColor Yellow

# Verify the settings
$crashControl = Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl"
Write-Host "Updated Crash Control Settings:" -ForegroundColor Green
Write-Host "  WriteEventToLog: $($crashControl.WriteEventToLog)" -ForegroundColor Cyan
Write-Host "  DumpFile: $($crashControl.DumpFile)" -ForegroundColor Cyan
Write-Host "  DumpType: $($crashControl.DumpType)" -ForegroundColor Cyan
Write-Host "  MinidumpDir: $($crashControl.MinidumpDir)" -ForegroundColor Cyan
Write-Host "  AutoReboot: $($crashControl.AutoReboot)" -ForegroundColor Cyan
Write-Host "  CrashDumpEnabled: $($crashControl.CrashDumpEnabled)" -ForegroundColor Cyan

# Check if minidump directory exists now
if (Test-Path $minidumpDir) {
    Write-Host "`n✓ Minidump directory created: $minidumpDir" -ForegroundColor Green
} else {
    Write-Host "`n✗ Failed to create minidump directory" -ForegroundColor Red
}

Write-Host "`nConfiguration updated successfully!" -ForegroundColor Green
Write-Host "You may need to restart the system for all changes to take effect." -ForegroundColor Yellow
Write-Host "`nNow test your BSOD application and check for:" -ForegroundColor Cyan
Write-Host '1. Event Log entries in Windows Logs > System (Event ID 41, 1001)' -ForegroundColor White
Write-Host '2. Minidump files in C:\Windows\Minidump' -ForegroundColor White
Write-Host '3. Full memory dump in C:\Windows\MEMORY.DMP' -ForegroundColor White

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
