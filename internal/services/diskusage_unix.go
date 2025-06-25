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
	blockSize := stat.Bsize
	//nolint:gosec // Safe conversion for disk space calculation
	totalSize := int64(stat.Blocks) * int64(blockSize)
	//nolint:gosec // Safe conversion for disk space calculation
	freeSize := int64(stat.Bavail) * int64(blockSize)

	return &DiskUsage{
		Free:  freeSize,
		Total: totalSize,
	}, nil
}