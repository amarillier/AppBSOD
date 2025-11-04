package main

/*
#cgo LDFLAGS: -lntdll -ladvapi32 -lwevtapi
#include <windows.h>
#include <winternl.h>
#include <sddl.h>
#include <winevt.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Minidump structures
#define MINIDUMP_SIGNATURE 0x504d444d  // 'MDMP'
#define MINIDUMP_VERSION 42899

typedef struct {
    ULONG32 Signature;
    ULONG32 Version;
    ULONG32 NumberOfStreams;
    ULONG32 StreamDirectoryRva;
    ULONG32 CheckSum;
    ULONG32 TimeDateStamp;
    ULONG64 Flags;
} MINIDUMP_HEADER;

typedef struct {
    ULONG32 StreamType;
    ULONG32 DataSize;
    ULONG32 Rva;
} MINIDUMP_DIRECTORY;

typedef struct {
    ULONG32 ThreadId;
    ULONG32 __alignment;
    ULONG64 ExceptionRecord;
    ULONG64 ContextRecord;
} MINIDUMP_EXCEPTION_STREAM;

typedef struct {
    ULONG32 ExceptionCode;
    ULONG32 ExceptionFlags;
    ULONG64 ExceptionRecord;
    ULONG64 ExceptionAddress;
    ULONG32 NumberParameters;
    ULONG32 __unusedAlignment;
    ULONG64 ExceptionInformation[15];
} MINIDUMP_EXCEPTION;

// For kernel dumps, bug check info is in ExceptionInformation[0-4]
// ExceptionCode contains the bug check code

// Bug check stream type
#define UnusedStream 0
#define SystemInfoStream 7
#define ExceptionStream 6

// Check if running as administrator
BOOL IsRunningAsAdmin() {
    BOOL isAdmin = FALSE;
    PSID adminGroup = NULL;
    SID_IDENTIFIER_AUTHORITY ntAuthority = SECURITY_NT_AUTHORITY;

    if (AllocateAndInitializeSid(&ntAuthority, 2, SECURITY_BUILTIN_DOMAIN_RID,
        DOMAIN_ALIAS_RID_ADMINS, 0, 0, 0, 0, 0, 0, &adminGroup)) {
        CheckTokenMembership(NULL, adminGroup, &isAdmin);
        FreeSid(adminGroup);
    }

    return isAdmin;
}

// Function to read minidump bug check info
int ReadMinidumpBugCheck(const char* dumpPath, ULONG32* bugCheckCode, ULONG32* param1, ULONG32* param2, ULONG32* param3, ULONG32* param4) {
    HANDLE hFile = CreateFileA(dumpPath, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, 0, NULL);
    if (hFile == INVALID_HANDLE_VALUE) {
        return 0;
    }

    DWORD fileSize = GetFileSize(hFile, NULL);
    if (fileSize == INVALID_FILE_SIZE || fileSize < sizeof(MINIDUMP_HEADER)) {
        CloseHandle(hFile);
        return 0;
    }

    HANDLE hMap = CreateFileMappingA(hFile, NULL, PAGE_READONLY, 0, 0, NULL);
    if (hMap == NULL) {
        CloseHandle(hFile);
        return 0;
    }

    LPVOID pView = MapViewOfFile(hMap, FILE_MAP_READ, 0, 0, 0);
    if (pView == NULL) {
        CloseHandle(hMap);
        CloseHandle(hFile);
        return 0;
    }

    MINIDUMP_HEADER* pHeader = (MINIDUMP_HEADER*)pView;
    if (pHeader->Signature != MINIDUMP_SIGNATURE) {
        UnmapViewOfFile(pView);
        CloseHandle(hMap);
        CloseHandle(hFile);
        return 0;
    }

    // Look for ExceptionStream (type 6) which contains bug check info
    MINIDUMP_DIRECTORY* pDir = (MINIDUMP_DIRECTORY*)((BYTE*)pView + pHeader->StreamDirectoryRva);

    for (ULONG i = 0; i < pHeader->NumberOfStreams; i++) {
        if (pDir[i].StreamType == ExceptionStream) {
            MINIDUMP_EXCEPTION_STREAM* pExStream = (MINIDUMP_EXCEPTION_STREAM*)((BYTE*)pView + pDir[i].Rva);

            // Check if ExceptionRecord is valid (non-zero means it points to an exception record)
            if (pExStream->ExceptionRecord != 0) {
                // ExceptionRecord is an RVA pointing to MINIDUMP_EXCEPTION
                MINIDUMP_EXCEPTION* pException = (MINIDUMP_EXCEPTION*)((BYTE*)pView + (ULONG32)pExStream->ExceptionRecord);

                // For kernel dumps, ExceptionCode is the bug check code
                // ExceptionInformation contains the parameters
                *bugCheckCode = pException->ExceptionCode;
                *param1 = (ULONG32)(pException->ExceptionInformation[0] & 0xFFFFFFFF);
                *param2 = (ULONG32)(pException->ExceptionInformation[1] & 0xFFFFFFFF);
                *param3 = (ULONG32)(pException->ExceptionInformation[2] & 0xFFFFFFFF);
                *param4 = (ULONG32)(pException->ExceptionInformation[3] & 0xFFFFFFFF);
            } else {
                // Try reading as if it's a direct bug check structure (legacy format)
                // Some dumps might have bug check info directly in the stream
                ULONG32* pData = (ULONG32*)((BYTE*)pView + pDir[i].Rva);
                *bugCheckCode = pData[0];
                *param1 = pData[1];
                *param2 = pData[2];
                *param3 = pData[3];
                *param4 = pData[4];
            }

            UnmapViewOfFile(pView);
            CloseHandle(hMap);
            CloseHandle(hFile);
            return 1;
        }
    }

    UnmapViewOfFile(pView);
    CloseHandle(hMap);
    CloseHandle(hFile);
    return 0;
}

// Function to get bug check info from Event Log (Event ID 1001)
// Returns 1 if found, 0 if not found
// Simplified version that gets the most recent Event ID 1001
int GetBugCheckFromEventLog(ULONG32* bugCheckCode, ULONG32* param1, ULONG32* param2, ULONG32* param3, ULONG32* param4) {
    EVT_HANDLE hResults = NULL;
    EVT_HANDLE hEvent = NULL;
    int found = 0;

    // Create XPath query for Event ID 1001 in System log
    LPCWSTR query = L"*[System[(EventID=1001)]]";
    LPCWSTR path = L"System";

    // Create query - get most recent events first
    hResults = EvtQuery(NULL, path, query, EvtQueryFilePath | EvtQueryReverseDirection);
    if (hResults == NULL) {
        return 0;
    }

    // Read the most recent event
    DWORD dwReturned = 0;
    EVT_HANDLE hEvents[1];

    if (EvtNext(hResults, 1, hEvents, 1000, 0, &dwReturned)) {
        if (dwReturned > 0) {
            hEvent = hEvents[0];

            PEVT_VARIANT pRenderedValues = NULL;
            DWORD dwBufferSize = 0;
            DWORD dwBufferUsed = 0;
            DWORD dwPropertyCount = 0;

            // Get required buffer size
            if (!EvtRender(NULL, hEvent, EvtRenderEventXml, dwBufferSize, pRenderedValues, &dwBufferUsed, &dwPropertyCount)) {
                DWORD dwError = GetLastError();
                if (dwError == ERROR_INSUFFICIENT_BUFFER) {
                    dwBufferSize = dwBufferUsed;
                    pRenderedValues = (PEVT_VARIANT)malloc(dwBufferSize);
                    if (pRenderedValues != NULL) {
                        if (EvtRender(NULL, hEvent, EvtRenderEventXml, dwBufferSize, pRenderedValues, &dwBufferUsed, &dwPropertyCount)) {
                            // Parse XML to find BugCheckCode and parameters
                            LPWSTR xml = (LPWSTR)pRenderedValues;

                            // Look for BugCheckCode in XML
                            LPWSTR codeStart = wcsstr(xml, L"<Data Name=\"BugCheckCode\">");
                            if (codeStart != NULL) {
                                codeStart += wcslen(L"<Data Name=\"BugCheckCode\">");
                                LPWSTR codeEnd = wcsstr(codeStart, L"</Data>");
                                if (codeEnd != NULL) {
                                    WCHAR save = *codeEnd;
                                    *codeEnd = 0;
                                    *bugCheckCode = (ULONG32)wcstoul(codeStart, NULL, 16);
                                    *codeEnd = save;

                                    // Look for parameters
                                    LPWSTR param1Start = wcsstr(xml, L"<Data Name=\"BugCheckParameter1\">");
                                    if (param1Start != NULL) {
                                        param1Start += wcslen(L"<Data Name=\"BugCheckParameter1\">");
                                        LPWSTR param1End = wcsstr(param1Start, L"</Data>");
                                        if (param1End != NULL) {
                                            WCHAR save = *param1End;
                                            *param1End = 0;
                                            *param1 = (ULONG32)wcstoul(param1Start, NULL, 16);
                                            *param1End = save;
                                        }
                                    }

                                    LPWSTR param2Start = wcsstr(xml, L"<Data Name=\"BugCheckParameter2\">");
                                    if (param2Start != NULL) {
                                        param2Start += wcslen(L"<Data Name=\"BugCheckParameter2\">");
                                        LPWSTR param2End = wcsstr(param2Start, L"</Data>");
                                        if (param2End != NULL) {
                                            WCHAR save = *param2End;
                                            *param2End = 0;
                                            *param2 = (ULONG32)wcstoul(param2Start, NULL, 16);
                                            *param2End = save;
                                        }
                                    }

                                    LPWSTR param3Start = wcsstr(xml, L"<Data Name=\"BugCheckParameter3\">");
                                    if (param3Start != NULL) {
                                        param3Start += wcslen(L"<Data Name=\"BugCheckParameter3\">");
                                        LPWSTR param3End = wcsstr(param3Start, L"</Data>");
                                        if (param3End != NULL) {
                                            WCHAR save = *param3End;
                                            *param3End = 0;
                                            *param3 = (ULONG32)wcstoul(param3Start, NULL, 16);
                                            *param3End = save;
                                        }
                                    }

                                    LPWSTR param4Start = wcsstr(xml, L"<Data Name=\"BugCheckParameter4\">");
                                    if (param4Start != NULL) {
                                        param4Start += wcslen(L"<Data Name=\"BugCheckParameter4\">");
                                        LPWSTR param4End = wcsstr(param4Start, L"</Data>");
                                        if (param4End != NULL) {
                                            WCHAR save = *param4End;
                                            *param4End = 0;
                                            *param4 = (ULONG32)wcstoul(param4Start, NULL, 16);
                                            *param4End = save;
                                        }
                                    }

                                    found = 1;
                                }
                            }

                            free(pRenderedValues);
                        }
                    }
                }
            }

            EvtClose(hEvents[0]);
        }
    }

    if (hResults != NULL) {
        EvtClose(hResults);
    }

    return found;
}
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type CrashDumpInfo struct {
	Path     string
	Name     string
	Size     int64
	Modified time.Time
}

type BugCheckInfo struct {
	Code   uint32
	Param1 uint32
	Param2 uint32
	Param3 uint32
	Param4 uint32
	String string
	Found  bool
}

// Bug check code to string mapping
var bugCheckStrings = map[uint32]string{
	0x0000000A: "IRQL_NOT_LESS_OR_EQUAL",
	0x0000001E: "KMODE_EXCEPTION_NOT_HANDLED",
	0x0000003B: "SYSTEM_SERVICE_EXCEPTION",
	0x0000007E: "SYSTEM_THREAD_EXCEPTION_NOT_HANDLED",
	0x0000007F: "UNEXPECTED_KERNEL_MODE_TRAP",
	0x000000D1: "DRIVER_IRQL_NOT_LESS_OR_EQUAL",
	0x000000EA: "THREAD_STUCK_IN_DEVICE_DRIVER",
	0xC0000022: "STATUS_ACCESS_DENIED",
	0xC0000420: "STATUS_ASSERTION_FAILURE",
	0xC0000005: "STATUS_ACCESS_VIOLATION",
	0x0000007A: "KERNEL_DATA_INPAGE_ERROR",
	0x0000009F: "DRIVER_POWER_STATE_FAILURE",
	0x000000BE: "ATTEMPTED_WRITE_TO_READONLY_MEMORY",
	0x000000CE: "PAGE_FAULT_IN_NONPAGED_AREA",
	0x000000D8: "DRIVER_USED_EXCESSIVE_PTES",
	0x000000F4: "CRITICAL_OBJECT_TERMINATION",
	0x000000FE: "BUGCODE_USB_DRIVER",
	0x00000109: "CRITICAL_STRUCTURE_CORRUPTION",
	0x0000010E: "VIDEO_TDR_TIMEOUT_DETECTED",
	0x00000124: "WHEA_UNCORRECTABLE_ERROR",
	0x00000127: "PAGE_FAULT_WITH_INTERRUPTS_OFF",
	0x00000133: "DPC_WATCHDOG_VIOLATION",
	0x00000139: "KERNEL_SECURITY_CHECK_FAILURE",
	0x00000141: "KERNEL_MODE_HEAP_CORRUPTION",
	0x00000142: "KERNEL_MODE_HEAP_CORRUPTION",
	0x00000143: "KERNEL_MODE_HEAP_CORRUPTION",
	0x0000014A: "KERNEL_THREAD_PRIORITY_FLOOR_VIOLATION",
	0x0000014B: "KERNEL_THREAD_PRIORITY_FLOOR_VIOLATION",
	0x0000014C: "KERNEL_THREAD_PRIORITY_FLOOR_VIOLATION",
}

func getBugCheckString(code uint32) string {
	if str, ok := bugCheckStrings[code]; ok {
		return str
	}
	return fmt.Sprintf("UNKNOWN_0x%08X", code)
}

func isRunningAsAdmin() bool {
	// Check via C function first
	result := C.IsRunningAsAdmin()
	if result != 0 {
		return true
	}

	// Fallback: Try to open a file that requires admin rights
	// This is a heuristic, not perfect but better than nothing
	testPath := "C:\\Windows\\System32\\config\\sam"
	if _, err := os.Open(testPath); err == nil {
		return true
	}
	return false
}

func findCrashDumps() ([]CrashDumpInfo, error) {
	var dumps []CrashDumpInfo

	// Check if running as administrator
	isAdmin := isRunningAsAdmin()

	// Check minidump directory
	minidumpDir := "C:\\Windows\\Minidump"

	// Check if directory exists first
	if dirInfo, err := os.Stat(minidumpDir); err == nil {
		if !dirInfo.IsDir() {
			// Path exists but isn't a directory
		} else {
			// Directory exists, try to read it
			entries, err := os.ReadDir(minidumpDir)
			if err != nil {
				// Directory exists but we can't read it - might be permissions issue
				// Continue anyway, we'll still check for MEMORY.DMP
				// Note: We'll warn about admin requirement in listDumps if needed
			} else {
				// Successfully read directory, look for .dmp files
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".dmp") {
						fullPath := filepath.Join(minidumpDir, entry.Name())
						if info, err := entry.Info(); err == nil {
							dumpInfo := CrashDumpInfo{
								Path:     fullPath,
								Name:     entry.Name(),
								Size:     info.Size(),
								Modified: info.ModTime(),
							}
							dumps = append(dumps, dumpInfo)
						}
					}
				}
			}
		}
	} else if !isAdmin {
		// Directory doesn't exist or can't be accessed
		// If we're not admin, this might be a permissions issue
		// (We'll warn about this in listDumps)
	}

	// Check for full memory dump
	memoryDump := "C:\\Windows\\MEMORY.DMP"
	if info, err := os.Stat(memoryDump); err == nil {
		dumpInfo := CrashDumpInfo{
			Path:     memoryDump,
			Name:     "MEMORY.DMP",
			Size:     info.Size(),
			Modified: info.ModTime(),
		}
		dumps = append(dumps, dumpInfo)
	}

	// Sort by modification time (newest first)
	sort.Slice(dumps, func(i, j int) bool {
		return dumps[i].Modified.After(dumps[j].Modified)
	})

	return dumps, nil
}

func analyzeDumpNative(dumpPath string) error {
	fmt.Printf("Analyzing crash dump: %s\n", dumpPath)
	fmt.Println(strings.Repeat("=", 80))

	// Get file info
	fileInfo, err := os.Stat(dumpPath)
	if err != nil {
		return fmt.Errorf("error accessing dump file: %v", err)
	}

	fmt.Printf("Dump File Information:\n")
	fmt.Printf("  Path: %s\n", dumpPath)
	fmt.Printf("  Size: %.2f MB\n", float64(fileInfo.Size())/(1024*1024))
	fmt.Printf("  Modified: %s\n", fileInfo.ModTime().Format(time.RFC3339))
	fmt.Println()

	// Try to read bug check info from dump file using C function
	cPath := C.CString(dumpPath)
	defer C.free(unsafe.Pointer(cPath))

	var bugCheckCode, param1, param2, param3, param4 C.ULONG32

	result := C.ReadMinidumpBugCheck(cPath, &bugCheckCode, &param1, &param2, &param3, &param4)
	if result != 0 {
		code := uint32(bugCheckCode)
		bugCheckStr := getBugCheckString(code)

		fmt.Printf("Bug Check Information (from dump file):\n")
		fmt.Printf("  Bug Check Code: 0x%08X\n", code)
		fmt.Printf("  Bug Check String: %s\n", bugCheckStr)
		fmt.Printf("  Parameter 1: 0x%08X\n", uint32(param1))
		fmt.Printf("  Parameter 2: 0x%08X\n", uint32(param2))
		fmt.Printf("  Parameter 3: 0x%08X\n", uint32(param3))
		fmt.Printf("  Parameter 4: 0x%08X\n", uint32(param4))
		fmt.Println()
		return nil
	}

	// Fallback: Try to parse manually as a last resort
	bugCheckInfo, err := parseDumpFileManually(dumpPath)
	if err == nil && bugCheckInfo.Found {
		fmt.Printf("Bug Check Information (parsed from dump):\n")
		fmt.Printf("  Bug Check Code: 0x%08X\n", bugCheckInfo.Code)
		fmt.Printf("  Bug Check String: %s\n", bugCheckInfo.String)
		fmt.Printf("  Parameter 1: 0x%08X\n", bugCheckInfo.Param1)
		fmt.Printf("  Parameter 2: 0x%08X\n", bugCheckInfo.Param2)
		fmt.Printf("  Parameter 3: 0x%08X\n", bugCheckInfo.Param3)
		fmt.Printf("  Parameter 4: 0x%08X\n", bugCheckInfo.Param4)
		fmt.Println()
		return nil
	}

	// Fallback: Try to get bug check info from Windows Event Log
	fmt.Println("Could not extract bug check information from dump file.")
	fmt.Println("Attempting to retrieve from Windows Event Log (Event ID 1001)...")
	fmt.Println()

	var eventBugCheckCode, eventParam1, eventParam2, eventParam3, eventParam4 C.ULONG32
	eventResult := C.GetBugCheckFromEventLog(&eventBugCheckCode, &eventParam1, &eventParam2, &eventParam3, &eventParam4)

	if eventResult != 0 {
		code := uint32(eventBugCheckCode)
		bugCheckStr := getBugCheckString(code)

		fmt.Printf("Bug Check Information (from Event Log):\n")
		fmt.Printf("  Bug Check Code: 0x%08X\n", code)
		fmt.Printf("  Bug Check String: %s\n", bugCheckStr)
		fmt.Printf("  Parameter 1: 0x%08X\n", uint32(eventParam1))
		fmt.Printf("  Parameter 2: 0x%08X\n", uint32(eventParam2))
		fmt.Printf("  Parameter 3: 0x%08X\n", uint32(eventParam3))
		fmt.Printf("  Parameter 4: 0x%08X\n", uint32(eventParam4))
		fmt.Println()
		fmt.Println("Note: This information was retrieved from the most recent Event ID 1001")
		fmt.Println("      in the Windows System Event Log. It may not match this specific dump.")
		fmt.Println()
		return nil
	}

	fmt.Println("Could not find bug check information in Event Log either.")
	fmt.Println("This may be a full memory dump or an incomplete minidump.")
	fmt.Println("Try using WinDbg for detailed analysis:")
	fmt.Printf("  windbg -z %s\n", dumpPath)
	fmt.Println("You can also use PowerShell for some basic event log queries, e.g.:")
	fmt.Println("Get-WinEvent -FilterHashtable @{LogName='System'; ID=41} -MaxEvents 5 |  Select-Object TimeCreated, Id, LevelDisplayName, Message")

	return nil
}

// Manual parsing as fallback - reads minidump header structure
func parseDumpFileManually(dumpPath string) (BugCheckInfo, error) {
	file, err := os.Open(dumpPath)
	if err != nil {
		return BugCheckInfo{}, err
	}
	defer file.Close()

	// Read minidump header
	var signature uint32
	if err := binary.Read(file, binary.LittleEndian, &signature); err != nil {
		return BugCheckInfo{}, err
	}

	// Check for minidump signature 'MDMP' (0x504d444d)
	if signature != 0x504d444d {
		return BugCheckInfo{}, fmt.Errorf("not a valid minidump file")
	}

	// Read version
	var version uint32
	binary.Read(file, binary.LittleEndian, &version)

	// Read number of streams
	var numStreams uint32
	binary.Read(file, binary.LittleEndian, &numStreams)

	// Read stream directory RVA
	var streamDirRVA uint32
	binary.Read(file, binary.LittleEndian, &streamDirRVA)

	// Skip to stream directory
	file.Seek(int64(streamDirRVA), 0)

	// Look for ExceptionStream (type 6)
	for i := uint32(0); i < numStreams; i++ {
		var streamType uint32
		var dataSize uint32
		var rva uint32

		binary.Read(file, binary.LittleEndian, &streamType)
		binary.Read(file, binary.LittleEndian, &dataSize)
		binary.Read(file, binary.LittleEndian, &rva)

		if streamType == 6 { // ExceptionStream
			// Seek to exception stream
			file.Seek(int64(rva), 0)

			// Read MINIDUMP_EXCEPTION_STREAM structure
			var threadId uint32
			var alignment uint32
			var exceptionRecordRVA uint64
			var contextRecordRVA uint64

			binary.Read(file, binary.LittleEndian, &threadId)
			binary.Read(file, binary.LittleEndian, &alignment)
			binary.Read(file, binary.LittleEndian, &exceptionRecordRVA)
			binary.Read(file, binary.LittleEndian, &contextRecordRVA)

			// If ExceptionRecord is valid, read the exception structure
			if exceptionRecordRVA != 0 {
				// Seek to the exception record
				file.Seek(int64(exceptionRecordRVA), 0)

				// Read MINIDUMP_EXCEPTION structure
				var exceptionCode uint32
				var exceptionFlags uint32
				var exceptionRecord uint64
				var exceptionAddress uint64
				var numberParameters uint32
				var unusedAlignment uint32
				var exceptionInformation [15]uint64

				binary.Read(file, binary.LittleEndian, &exceptionCode)
				binary.Read(file, binary.LittleEndian, &exceptionFlags)
				binary.Read(file, binary.LittleEndian, &exceptionRecord)
				binary.Read(file, binary.LittleEndian, &exceptionAddress)
				binary.Read(file, binary.LittleEndian, &numberParameters)
				binary.Read(file, binary.LittleEndian, &unusedAlignment)
				binary.Read(file, binary.LittleEndian, &exceptionInformation)

				var info BugCheckInfo
				info.Code = exceptionCode
				info.Param1 = uint32(exceptionInformation[0] & 0xFFFFFFFF)
				info.Param2 = uint32(exceptionInformation[1] & 0xFFFFFFFF)
				info.Param3 = uint32(exceptionInformation[2] & 0xFFFFFFFF)
				info.Param4 = uint32(exceptionInformation[3] & 0xFFFFFFFF)
				info.String = getBugCheckString(info.Code)
				info.Found = true

				return info, nil
			} else {
				// Fallback: try reading bug check codes directly from stream start
				// (some older dump formats might have this)
				file.Seek(int64(rva), 0)
				var info BugCheckInfo
				binary.Read(file, binary.LittleEndian, &info.Code)
				binary.Read(file, binary.LittleEndian, &info.Param1)
				binary.Read(file, binary.LittleEndian, &info.Param2)
				binary.Read(file, binary.LittleEndian, &info.Param3)
				binary.Read(file, binary.LittleEndian, &info.Param4)

				// Only consider it valid if the code looks like a bug check code
				// (typically 0x00000000 to 0xFFFFFFFF, but not 0)
				if info.Code != 0 {
					info.String = getBugCheckString(info.Code)
					info.Found = true
					return info, nil
				}
			}
		}
	}

	return BugCheckInfo{}, fmt.Errorf("exception stream not found")
}

func listDumps() {
	// Check if running as administrator
	isAdmin := isRunningAsAdmin()

	// Check minidump directory first
	minidumpDir := "C:\\Windows\\Minidump"
	fmt.Printf("Checking for crash dumps in:\n")
	fmt.Printf("  - %s\n", minidumpDir)
	fmt.Printf("  - C:\\Windows\\MEMORY.DMP\n")

	if !isAdmin {
		fmt.Printf("\n⚠ Warning: Not running as Administrator\n")
		fmt.Printf("  Reading minidump files may require elevated permissions.\n")
		fmt.Printf("  Run as Administrator for best results.\n\n")
	} else {
		fmt.Println()
	}

	dumps, err := findCrashDumps()
	if err != nil {
		fmt.Printf("Error finding crash dumps: %v\n", err)
		return
	}

	if len(dumps) == 0 {
		fmt.Println("No crash dumps found")
		if !isAdmin {
			fmt.Printf("\nNote: Make sure you have permissions to read:\n")
			fmt.Printf("  - %s (requires Administrator)\n", minidumpDir)
			fmt.Printf("  - C:\\Windows\\MEMORY.DMP\n")
			fmt.Printf("\nTry running as Administrator if minidump directory exists.\n")
		} else {
			fmt.Printf("\nNote: Make sure the directories exist:\n")
			fmt.Printf("  - %s\n", minidumpDir)
			fmt.Printf("  - C:\\Windows\\MEMORY.DMP\n")
		}
		return
	}

	fmt.Printf("Found %d crash dump(s):\n\n", len(dumps))
	for i, dump := range dumps {
		fmt.Printf("%d. %s\n", i+1, dump.Name)
		fmt.Printf("   Path: %s\n", dump.Path)
		fmt.Printf("   Size: %.2f MB\n", float64(dump.Size)/(1024*1024))
		fmt.Printf("   Modified: %s\n", dump.Modified.Format(time.RFC3339))
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Crash Dump Analyzer - List and analyze Windows crash dumps")
		fmt.Println("\nUsage:")
		fmt.Println("  CrashDumpAnalyzer.exe list")
		fmt.Println("  CrashDumpAnalyzer.exe analyze <dump_file_path>")
		fmt.Println("  CrashDumpAnalyzer.exe analyze <index>  # Analyze by index from list")
		fmt.Println("\nExamples:")
		fmt.Println("  CrashDumpAnalyzer.exe list")
		fmt.Println("  CrashDumpAnalyzer.exe analyze C:\\Windows\\Minidump\\01234567-01.dmp")
		fmt.Println("  CrashDumpAnalyzer.exe analyze 1")
		return
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "list":
		listDumps()
	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please specify a dump file path or index")
			return
		}

		arg := os.Args[2]

		// Check if it's a numeric index
		if index, err := strconv.Atoi(arg); err == nil {
			dumps, err := findCrashDumps()
			if err != nil || index < 1 || index > len(dumps) {
				fmt.Printf("Error: Invalid dump index\n")
				return
			}
			analyzeDumpNative(dumps[index-1].Path)
		} else {
			// Treat as file path
			if _, err := os.Stat(arg); err != nil {
				fmt.Printf("Error: Dump file not found: %s\n", arg)
				return
			}
			analyzeDumpNative(arg)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
