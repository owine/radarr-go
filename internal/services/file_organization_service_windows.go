//go:build windows

package services

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceW")
)

// getDiskSpace returns available disk space in bytes for Windows systems
func (s *FileOrganizationService) getDiskSpace(path string) (int64, error) {
	// Convert path to absolute and get drive letter
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Extract drive letter (e.g., "C:\")
	drive := filepath.VolumeName(absPath)
	if drive == "" {
		drive = absPath[:3] // fallback for "C:\" format
	}

	// Convert to UTF-16 for Windows API
	drivePtr, err := syscall.UTF16PtrFromString(drive + "\\")
	if err != nil {
		return 0, fmt.Errorf("failed to convert drive path: %w", err)
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	// Call GetDiskFreeSpaceW
	ret, _, err := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(drivePtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return 0, fmt.Errorf("GetDiskFreeSpaceW failed: %w", err)
	}

	// Convert to int64, checking for overflow
	if freeBytesAvailable > uint64(int64(^uint64(0)>>1)) {
		return int64(^uint64(0) >> 1), nil // Return MaxInt64
	}

	return int64(freeBytesAvailable), nil // #nosec G115 - overflow checked above
}
