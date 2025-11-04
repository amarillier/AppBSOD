package main

/*
#cgo LDFLAGS: -lntdll
#include <windows.h>
#include <winternl.h>

void __stdcall RtlAdjustPrivilege(
    ULONG Privilege,
    BOOLEAN Enable,
    BOOLEAN CurrentThread,
    PBOOLEAN Enabled);

NTSTATUS __stdcall NtRaiseHardError(
    NTSTATUS ErrorStatus,
    ULONG NumberOfParameters,
    ULONG UnicodeStringParameterMask,
    PULONG_PTR Parameters,
    ULONG ValidResponseOptions,
    PULONG Response);
*/
import "C"

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	SE_SHUTDOWN_PRIVILEGE  = C.ULONG(19)
	OPTION_SHUTDOWN_SYSTEM = C.ULONG(6)

	STATUS_ACCESS_DENIED     = C.NTSTATUS(-1073741790) // 0xC0000022
	STATUS_ASSERTION_FAILURE = C.NTSTATUS(-1073741024) // 0xC0000420
)

// Bug check code to string mapping (common codes)
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
}

// Function to get bug check string from code
func getBugCheckString(ntstatus uint32) string {
	if str, ok := bugCheckStrings[ntstatus]; ok {
		return str
	}
	return fmt.Sprintf("UNKNOWN_0x%08X", ntstatus)
}

func parseNTStatus(arg string) (C.NTSTATUS, error) {
	switch strings.ToLower(arg) {
	case "access_denied":
		return STATUS_ACCESS_DENIED, nil
	case "assertion_failure":
		return STATUS_ASSERTION_FAILURE, nil
	default:
		// Try to parse custom hex code
		if strings.HasPrefix(arg, "0x") {
			val, err := strconv.ParseUint(arg[2:], 16, 32)
			if err != nil {
				return 0, fmt.Errorf("invalid hex NTSTATUS code")
			}
			return C.NTSTATUS(int32(val)), nil
		}
		return 0, fmt.Errorf("unknown NTSTATUS code or keyword")
	}
}

func forceSystemFlush() {
	// Give system time to process any pending operations
	time.Sleep(200 * time.Millisecond)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("BusinessAppBSOD Enhanced - Trigger a Blue Screen of Death (BSOD) on Windows")
		fmt.Println("Warning: This will crash your system immediately after providing a parameter!")
		fmt.Println("\nUsage: BusinessAppBSOD_goX.exe <access_denied | assertion_failure | 0xC0000005> [--bug-check-string]")
		fmt.Println("BusinessAppBSOD_goX.exe access_denied")
		fmt.Println("BusinessAppBSOD_goX.exe assertion_failure --bug-check-string")
		fmt.Println("BusinessAppBSOD_goX.exe 0xC0000005   # STATUS_ACCESS_VIOLATION")
		fmt.Println("BusinessAppBSOD_goX.exe 0xDEADBEEF --bug-check-string   # Custom code with bug check string")

		return
	}

	ntstatus, err := parseNTStatus(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Check for --bug-check-string flag
	includeBugCheckString := false
	for _, arg := range os.Args[2:] {
		if arg == "--bug-check-string" {
			includeBugCheckString = true
			break
		}
	}

	var enabled C.BOOLEAN
	var response C.ULONG

	fmt.Println("Adjusting privilege...")
	C.RtlAdjustPrivilege(SE_SHUTDOWN_PRIVILEGE, C.BOOLEAN(1), C.BOOLEAN(0), &enabled)

	// Log event before triggering BSOD
	eventMessage := fmt.Sprintf("BusinessAppBSOD Enhanced triggering BSOD with NTSTATUS: 0x%x", uint32(ntstatus))
	if includeBugCheckString {
		bugCheckStr := getBugCheckString(uint32(ntstatus))
		eventMessage += fmt.Sprintf(" (Bug Check: %s)", bugCheckStr)
	}
	fmt.Println("Logging event:", eventMessage)

	// Force system to flush buffers
	fmt.Println("Flushing system buffers...")
	forceSystemFlush()

	fmt.Printf("Triggering BSOD with NTSTATUS: 0x%x\n", uint32(ntstatus))
	if includeBugCheckString {
		bugCheckStr := getBugCheckString(uint32(ntstatus))
		fmt.Printf("Bug Check String: %s\n", bugCheckStr)
	}

	// Force Go runtime to flush output
	os.Stdout.Sync()

	// Give a moment for logging to complete
	time.Sleep(500 * time.Millisecond)

	C.NtRaiseHardError(
		ntstatus,
		0,
		0,
		nil,
		OPTION_SHUTDOWN_SYSTEM,
		&response,
	)
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
