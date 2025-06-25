// +build windows

package services

import (
	"syscall"
	"unsafe"
)

// getDiskUsageForPath returns disk usage information for Windows systems
func getDiskUsageForPath(path string) (*DiskUsage, error) {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes int64

	r1, _, err := c.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if r1 == 0 {
		return nil, err
	}

	return &DiskUsage{
		Free:  freeBytesAvailable,
		Total: totalNumberOfBytes,
	}, nil
}