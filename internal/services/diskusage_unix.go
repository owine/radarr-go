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
	//nolint:gosec // Safe conversion: filesystem block sizes are always positive (G115 on Linux)
	blockSize := uint64(stat.Bsize)
	//nolint:gosec,unconvert // Safe conversion: filesystem values are always positive, no overflow risk (G115)
	totalSize := int64(uint64(stat.Blocks) * blockSize)
	//nolint:gosec,unconvert // Safe conversion: filesystem values are always positive, no overflow risk (G115)
	freeSize := int64(uint64(stat.Bavail) * blockSize)

	return &DiskUsage{
		Free:  freeSize,
		Total: totalSize,
	}, nil
}
