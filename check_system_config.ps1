# PowerShell script to check Windows BSOD logging configuration
Write-Host "Checking Windows BSOD logging configuration..." -ForegroundColor Green

# Check if system is configured to write events to system log
$startupRecovery = Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" -ErrorAction SilentlyContinue
if ($startupRecovery) {
    Write-Host "Crash Control Settings:" -ForegroundColor Yellow
    Write-Host "  WriteEventToLog: $($startupRecovery.WriteEventToLog)" -ForegroundColor $(if ($startupRecovery.WriteEventToLog -eq 1) { "Green" } else { "Red" })
    Write-Host "  DumpFile: $($startupRecovery.DumpFile)" -ForegroundColor Cyan
    Write-Host "  DumpType: $($startupRecovery.DumpType)" -ForegroundColor $(if ($startupRecovery.DumpType -gt 0) { "Green" } else { "Red" })
    Write-Host "  MinidumpDir: $($startupRecovery.MinidumpDir)" -ForegroundColor Cyan
    Write-Host "  AutoReboot: $($startupRecovery.AutoReboot)" -ForegroundColor Cyan
    Write-Host "  CrashDumpEnabled: $($startupRecovery.CrashDumpEnabled)" -ForegroundColor Cyan
    
    # Check for configuration issues
    $issues = @()
    if ($startupRecovery.WriteEventToLog -ne 1) { $issues += "WriteEventToLog should be 1" }
    if ($startupRecovery.DumpType -eq $null -or $startupRecovery.DumpType -eq 0) { $issues += "DumpType should be 1-7" }
    
    if ($issues.Count -gt 0) {
        Write-Host "`n⚠ Configuration Issues Found:" -ForegroundColor Red
        foreach ($issue in $issues) {
            Write-Host "  - $issue" -ForegroundColor Red
        }
        Write-Host "`nRun .\fix_bsod_config.ps1 as Administrator to fix these issues" -ForegroundColor Yellow
    } else {
        Write-Host "`n✓ BSOD logging configuration looks good" -ForegroundColor Green
    }
} else {
    Write-Host "Could not read crash control settings" -ForegroundColor Red
}

# Check event log settings
Write-Host "`nChecking Event Log Settings..." -ForegroundColor Yellow
$systemLog = Get-WinEvent -ListLog "System" -ErrorAction SilentlyContinue
if ($systemLog) {
    Write-Host "System Log Status: $($systemLog.IsEnabled)" -ForegroundColor Cyan
    Write-Host "System Log Max Size: $($systemLog.MaximumSizeInBytes) bytes" -ForegroundColor Cyan
}

# Check for recent BSOD events
Write-Host "`nChecking for recent BSOD events..." -ForegroundColor Yellow
try {
    $bsodEvents = Get-WinEvent -FilterHashtable @{LogName='System'; ID=41} -MaxEvents 5 -ErrorAction SilentlyContinue
    if ($bsodEvents) {
        Write-Host "Found $($bsodEvents.Count) recent system crash events:" -ForegroundColor Green
        foreach ($event in $bsodEvents) {
            Write-Host "  $($event.TimeCreated) - ID: $($event.Id)" -ForegroundColor Cyan
        }
    } else {
        Write-Host "No recent BSOD events found" -ForegroundColor Yellow
    }
} catch {
    Write-Host "Could not query event log: $($_.Exception.Message)" -ForegroundColor Red
}

# Check for minidump files
Write-Host "`nChecking for minidump files..." -ForegroundColor Yellow
$minidumpPath = "C:\Windows\Minidump"
if (Test-Path $minidumpPath) {
    $dumps = Get-ChildItem $minidumpPath -Filter "*.dmp" | Sort-Object LastWriteTime -Descending | Select-Object -First 5
    if ($dumps) {
        Write-Host "Found $($dumps.Count) recent minidump files:" -ForegroundColor Green
        foreach ($dump in $dumps) {
            Write-Host "  $($dump.Name) - $($dump.LastWriteTime)" -ForegroundColor Cyan
        }
    } else {
        Write-Host "No minidump files found" -ForegroundColor Yellow
    }
} else {
    Write-Host "Minidump directory not found" -ForegroundColor Red
}

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
