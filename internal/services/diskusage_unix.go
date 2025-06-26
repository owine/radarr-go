//go:build !windows
// +build !windows

package services

import (
	"syscall"
)

// getDiskUsageForPath returns disk usage information for Unix-like systems
func getDiskUsageForPath(path string) (*DiskUsage, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, err
	}

	// Calculate sizes in bytes
	// Note: stat.Bsize is int64 on Linux, uint32 on Darwin, uint64 on FreeBSD
	blockSize := uint64(stat.Bsize) //#nosec G115 // Safe conversion: filesystem block sizes are always positive
	totalSize := int64(stat.Blocks * blockSize)         //#nosec G115 // Blocks is uint64 on all platforms
	//nolint:unconvert // uint64 conversion needed for FreeBSD compatibility (int64->uint64)
	freeSize := int64(uint64(stat.Bavail) * blockSize) //#nosec G115 // Bavail is int64 on FreeBSD, uint64 on Linux/Darwin

	return &DiskUsage{
		Free:  freeSize,
		Total: totalSize,
	}, nil
}
