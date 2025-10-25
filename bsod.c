#include <windows.h>
#include <winternl.h>
#include <stdio.h>
#include <stdlib.h>

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
    printf("\nUsage: BusinessAppBSOD.exe <access_denied | assertion_failure | 0xC0000005>\n");
    printf("BusinessAppBSOD.exe access_denied\n");
    printf("BusinessAppBSOD.exe assertion_failure\n");
    printf("BusinessAppBSOD.exe 0xC0000005   # STATUS_ACCESS_VIOLATION\n");
    printf("BusinessAppBSOD.exe 0xDEADBEEF   # Custom code\n");
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        print_usage();
        return 1;
    }

    NTSTATUS ntstatus = parse_ntstatus(argv[1]);
    BOOLEAN enabled;
    ULONG response;

    printf("Adjusting privilege...\n");
    RtlAdjustPrivilege(SE_SHUTDOWN_PRIVILEGE, TRUE, FALSE, &enabled);

    printf("Triggering BSOD with NTSTATUS: 0x%x\n", (unsigned int)ntstatus);
    
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
