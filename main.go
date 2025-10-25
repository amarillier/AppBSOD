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
)

const (
	SE_SHUTDOWN_PRIVILEGE  = C.ULONG(19)
	OPTION_SHUTDOWN_SYSTEM = C.ULONG(6)

	STATUS_ACCESS_DENIED     = C.NTSTATUS(-1073741790) // 0xC0000022
	STATUS_ASSERTION_FAILURE = C.NTSTATUS(-1073741024) // 0xC0000420
)

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("BusinessAppBSOD - Trigger a Blue Screen of Death (BSOD) on Windows")
		fmt.Println("Warning: This will crash your system immediately after providing a parameter!")
		fmt.Println("\nUsage: BusinessAppBSOD.exe <access_denied | assertion_failure | 0xC0000005>")
		fmt.Println("BusinessAppBSOD.exe access_denied")
		fmt.Println("BusinessAppBSOD.exe assertion_failure")
		fmt.Println("BusinessAppBSOD.exe 0xC0000005   # STATUS_ACCESS_VIOLATION")
		fmt.Println("BusinessAppBSOD.exe 0xDEADBEEF   # Custom code")

		return
	}

	ntstatus, err := parseNTStatus(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var enabled C.BOOLEAN
	var response C.ULONG

	fmt.Println("Adjusting privilege...")
	C.RtlAdjustPrivilege(SE_SHUTDOWN_PRIVILEGE, C.BOOLEAN(1), C.BOOLEAN(0), &enabled)

	fmt.Printf("Triggering BSOD with NTSTATUS: 0x%x\n", uint32(ntstatus))
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
