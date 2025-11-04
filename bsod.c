#include <windows.h>
#include <winternl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Function declarations
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

// Constants
#define SE_SHUTDOWN_PRIVILEGE 19
#define OPTION_SHUTDOWN_SYSTEM 6
#define STATUS_ACCESS_DENIED ((NTSTATUS)0xC0000022L)
#define STATUS_ASSERTION_FAILURE ((NTSTATUS)0xC0000420L)

// Bug check code to string mapping (common codes)
typedef struct {
    ULONG code;
    const char* string;
} BugCheckString;

BugCheckString bugCheckStrings[] = {
    {0x0000000A, "IRQL_NOT_LESS_OR_EQUAL"},
    {0x0000001E, "KMODE_EXCEPTION_NOT_HANDLED"},
    {0x0000003B, "SYSTEM_SERVICE_EXCEPTION"},
    {0x0000007E, "SYSTEM_THREAD_EXCEPTION_NOT_HANDLED"},
    {0x0000007F, "UNEXPECTED_KERNEL_MODE_TRAP"},
    {0x000000D1, "DRIVER_IRQL_NOT_LESS_OR_EQUAL"},
    {0x000000EA, "THREAD_STUCK_IN_DEVICE_DRIVER"},
    {0xC0000022, "STATUS_ACCESS_DENIED"},
    {0xC0000420, "STATUS_ASSERTION_FAILURE"},
    {0xC0000005, "STATUS_ACCESS_VIOLATION"},
    {0, NULL}  // Sentinel
};

// Function to get bug check string from code
const char* get_bug_check_string(NTSTATUS ntstatus) {
    ULONG code = (ULONG)ntstatus;
    for (int i = 0; bugCheckStrings[i].string != NULL; i++) {
        if (bugCheckStrings[i].code == code) {
            return bugCheckStrings[i].string;
        }
    }
    return NULL;  // Unknown code
}

NTSTATUS parse_ntstatus(const char* arg) {
    if (strcmp(arg, "access_denied") == 0) {
        return STATUS_ACCESS_DENIED;
    } else if (strcmp(arg, "assertion_failure") == 0) {
        return STATUS_ASSERTION_FAILURE;
    } else if (strncmp(arg, "0x", 2) == 0) {
        return (NTSTATUS)strtoul(arg, NULL, 16);
    } else {
        return (NTSTATUS)strtoul(arg, NULL, 16);
    }
}

void print_usage() {
    printf("BusinessAppBSOD - Trigger a Blue Screen of Death (BSOD) on Windows\n");
    printf("Warning: This will crash your system immediately after providing a parameter!\n");
    printf("\nUsage: BusinessAppBSOD.exe <access_denied | assertion_failure | 0xC0000005> [--bug-check-string]\n");
    printf("BusinessAppBSOD.exe access_denied\n");
    printf("BusinessAppBSOD.exe assertion_failure --bug-check-string\n");
    printf("BusinessAppBSOD.exe 0xC0000005   # STATUS_ACCESS_VIOLATION\n");
    printf("BusinessAppBSOD.exe 0xDEADBEEF --bug-check-string   # Custom code with bug check string\n");
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        print_usage();
        return 1;
    }

    NTSTATUS ntstatus = parse_ntstatus(argv[1]);
    BOOLEAN enabled;
    ULONG response;
    
    // Check for --bug-check-string flag
    int include_bug_check_string = 0;
    for (int i = 2; i < argc; i++) {
        if (strcmp(argv[i], "--bug-check-string") == 0) {
            include_bug_check_string = 1;
            break;
        }
    }

    printf("Adjusting privilege...\n");
    RtlAdjustPrivilege(SE_SHUTDOWN_PRIVILEGE, TRUE, FALSE, &enabled);

    printf("Triggering BSOD with NTSTATUS: 0x%x\n", (unsigned int)ntstatus);
    
    // Display bug check string if requested
    if (include_bug_check_string) {
        const char* bug_check_str = get_bug_check_string(ntstatus);
        if (bug_check_str) {
            printf("Bug Check String: %s\n", bug_check_str);
        } else {
            printf("Bug Check String: UNKNOWN_0x%08X\n", (unsigned int)ntstatus);
        }
    }
    
    // Force a flush to ensure output is written before crash
    fflush(stdout);
    
    NtRaiseHardError(
        ntstatus,
        0,
        0,
        NULL,
        OPTION_SHUTDOWN_SYSTEM,
        &response
    );

    return 0;
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
