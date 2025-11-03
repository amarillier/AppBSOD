# BusinessAppBSOD

A Windows application designed to trigger Blue Screen of Death (BSOD) for testing purposes. This project includes both Go and C implementations to compare event logging behavior.

⚠️ **WARNING**: This application **will** immediately crash your system! Only use on virtual machines or test systems in controlled environments.

## Project Overview

This includes both C and Go applications using `NtRaiseHardError`, because Go protections might not create proper Windows event log entries, while the C implementation bypasses Go runtime limitations and causes true BSOD events. Note: The Go version does also cause a true BSOD, but not all the same logging.

## Files

### Source Code
- `main.go` - Go implementation using CGO
- `main_enhanced.go` - Enhanced Go version with explicit event logging
- `bsod.c` - Pure C implementation
- `windows.h.sample` - Sample Windows header for cross-compilation on MacOS
- `winternl.h.sample` - Sample Winternl header for cross-compilation on MacOS

### Compilation Scripts
- `compile_c.sh` - Bash script to compile C version
- `compile_go.sh` - Bash script to compile Go version
- `compile_all.sh` - Bash script to compile both versions

- To sync to, and build on Windows:
- `sync2windows.sh` - Sync files between Mac and Windows
- `sync2windows.sh sync-back` - Sync files back from Windows to Mac
- `compile_go.ps1` - PowerShell script to compile Go version
- `compile_c.ps1` - PowerShell script to compile C version
- `compile_all.ps1` - PowerShell script to compile both versions


### Configuration & Analysis
- `check_system_config.ps1` - Check Windows BSOD logging configuration
- `fix_bsod_config.ps1` - Fix Windows BSOD logging configuration to enable all event logging (run as Administrator)


## Prerequisites

### For Go Version
- Go 1.25.1 or later
- MinGW-w64 GCC for Windows cross-compilation

### For C Version
- MinGW-w64 GCC (x86_64-w64-mingw32-gcc)
- Windows SDK (for headers)

## Compilation

### On Windows
```powershell
# Compile Go version (standard)
.\compile_go.ps1

# Compile Go version (both standard and enhanced)
.\compile_go.ps1 -Mode enhanced

# Compile C version
.\compile_c.ps1

# Compile both versions
.\compile_all.ps1
```

### On Mac/Linux (Cross-compilation)
```bash
# Compile Go version (standard)
./compile_go.sh

# Compile Go version (both standard and enhanced)
./compile_go.sh enhanced

# Compile C version
./compile_c.sh
```

## Usage

### Go Version (Standard)
```powershell
.\BusinessAppBSOD_go.exe <access_denied | assertion_failure | 0xC0000005>
```

### Go Version (Enhanced - includes additional logging)
```powershell
.\BusinessAppBSOD_goX.exe <access_denied | assertion_failure | 0xC0000005>
```

### C Version
```powershell
.\BusinessAppBSOD.exe <access_denied | assertion_failure | 0xC0000005>
```

### Available NTSTATUS Codes
- `access_denied` - STATUS_ACCESS_DENIED (0xC0000022)
- `assertion_failure` - STATUS_ASSERTION_FAILURE (0xC0000420)
- `0xC0000005` - STATUS_ACCESS_VIOLATION
- `0xDEADBEEF` - Custom code (example)

## Windows Event Log Analysis

### 1. Using Event Viewer (GUI)

1. **Open Event Viewer**:
   - Press `Windows + R`, type `eventvwr.msc`, press Enter
   - Or search "Event Viewer" in Start menu

2. **Navigate to System Log**:
   - Expand "Windows Logs"
   - Click on "System"

3. **Filter for BSOD Events**:
   - Right-click "System" → "Filter Current Log..."
   - Set Event IDs: `41, 1001`
   - Click "OK"

4. **Look for These Event IDs**:
   - **Event ID 41**: System rebooted without clean shutdown
   - **Event ID 1001**: Windows Error Reporting events

### 2. Using PowerShell

#### Check Recent BSOD Events
```powershell
# Get recent system crash events (Event ID 41)
Get-WinEvent -FilterHashtable @{LogName='System'; ID=41} -MaxEvents 10

# Get Windows Error Reporting events (Event ID 1001)
Get-WinEvent -FilterHashtable @{LogName='System'; ID=1001} -MaxEvents 10

# Get all BSOD-related events from last 24 hours
$startTime = (Get-Date).AddDays(-1)
Get-WinEvent -FilterHashtable @{LogName='System'; StartTime=$startTime} | 
    Where-Object {$_.Id -in @(41, 1001, 6008, 6009)}
```

#### Detailed Event Analysis
```powershell
# Get detailed information about crash events
$events = Get-WinEvent -FilterHashtable @{LogName='System'; ID=41} -MaxEvents 5
foreach ($event in $events) {
    Write-Host "Time: $($event.TimeCreated)" -ForegroundColor Yellow
    Write-Host "Event ID: $($event.Id)" -ForegroundColor Cyan
    Write-Host "Level: $($event.LevelDisplayName)" -ForegroundColor Green
    Write-Host "Message: $($event.Message)" -ForegroundColor White
    Write-Host "---" -ForegroundColor Gray
}
```

#### Search for Specific Error Codes
```powershell
# Search for specific NTSTATUS codes in event logs
Get-WinEvent -FilterHashtable @{LogName='System'} | 
    Where-Object {$_.Message -match "0xC0000022|0xC0000420|0xC0000005"}
```

### 3. Check Memory Dump Files

#### Using PowerShell
```powershell
# Check for minidump files
$minidumpDir = "C:\Windows\Minidump"
if (Test-Path $minidumpDir) {
    $dumps = Get-ChildItem $minidumpDir -Filter "*.dmp" | Sort-Object LastWriteTime -Descending
    Write-Host "Found $($dumps.Count) minidump files:" -ForegroundColor Green
    foreach ($dump in $dumps) {
        Write-Host "  $($dump.Name) - $($dump.LastWriteTime)" -ForegroundColor Cyan
    }
} else {
    Write-Host "No minidump directory found" -ForegroundColor Red
}

# Check for full memory dump
$memoryDump = "C:\Windows\MEMORY.DMP"
if (Test-Path $memoryDump) {
    $dumpInfo = Get-Item $memoryDump
    Write-Host "Full memory dump found: $($dumpInfo.Name)" -ForegroundColor Green
    Write-Host "Size: $([math]::Round($dumpInfo.Length / 1MB, 2)) MB" -ForegroundColor Cyan
    Write-Host "Last modified: $($dumpInfo.LastWriteTime)" -ForegroundColor Cyan
} else {
    Write-Host "No full memory dump found" -ForegroundColor Yellow
}
```

## System Configuration

### Check Current Configuration
```powershell
.\check_system_config.ps1
```

### Fix Configuration (Run as Administrator)
```powershell
.\fix_bsod_config.ps1
```

This script will:
- Enable event logging (`WriteEventToLog = 1`)
- Enable small memory dumps (`DumpType = 1`)
- Create minidump directory
- Enable automatic restart after BSOD

## Troubleshooting

### No Event Log Entries
1. **Check system configuration**:
   ```powershell
   .\check_system_config.ps1
   ```

2. **Fix configuration** (as Administrator):
   ```powershell
   .\fix_bsod_config.ps1
   ```

3. **Restart the system** after configuration changes

### No Memory Dumps
1. **Check page file size** - should be at least 1GB
2. **Verify disk space** - need space for dump files
3. **Check registry settings**:
   ```powershell
   Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl"
   ```

### Go vs C Differences
- **Go version**: May not create event logs due to Go language runtime protections (interference). The GoX version **should** log events
- **C version**: More reliable for event logging and memory dumps
- **Recommendation**: Use C version for testing event logging behavior

## Analysis Tools

### Windows Built-in Tools
- **Event Viewer** (`eventvwr.msc`)
- **BlueScreenView** (NirSoft) - Analyze minidump files
- **WinDbg** - Advanced crash dump analysis

### PowerShell Commands
```powershell
# Quick BSOD event summary
Get-WinEvent -FilterHashtable @{LogName='System'; ID=41} -MaxEvents 5 | 
    Select-Object TimeCreated, Id, LevelDisplayName, Message

# Check system crash configuration
Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\CrashControl" | 
    Select-Object WriteEventToLog, DumpType, MinidumpDir, DumpFile
```

## Safety Notes

- ⚠️ **Only use on virtual machines or test systems**
- ⚠️ **Save your work before running**
- ⚠️ **System will restart immediately after BSOD**
- ⚠️ **May cause data loss if not properly prepared**

## License

This project is for educational and testing purposes only. Use at your own risk.

## Contributing

This is a testing tool. Contributions should focus on improving analysis capabilities or documentation.

## "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942