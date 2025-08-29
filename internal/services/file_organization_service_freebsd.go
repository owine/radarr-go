//go:build freebsd
// +build freebsd

package services

import (
	"math"

	"golang.org/x/sys/unix"
)

// getDiskSpace returns available disk space in bytes for FreeBSD systems
func (s *FileOrganizationService) getDiskSpace(path string) (int64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, err
	}

	// Calculate available space with overflow protection
	bavail := uint64(stat.Bavail) // int64 on FreeBSD, convert to uint64
	bsize := uint64(stat.Bsize)   // uint64 on FreeBSD

	// Check for overflow in multiplication
	if bavail > 0 && bsize > 0 && bavail > uint64(math.MaxInt64)/bsize {
		// Handle overflow by returning MaxInt64 as available space
		return math.MaxInt64, nil
	}

	// Safe conversion after overflow check
	product := bavail * bsize // #nosec G115 - overflow checked above
	if product > math.MaxInt64 {
		return math.MaxInt64, nil // #nosec G115 - capped at MaxInt64
	}

	// Safe conversion - already checked that product <= math.MaxInt64 above
	return int64(product), nil // #nosec G115 - overflow checked above
}
