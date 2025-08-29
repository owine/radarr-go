//go:build unix

package services

import (
	"math"
	"syscall"
)

// getDiskSpace returns available disk space in bytes for Unix systems
func (s *FileOrganizationService) getDiskSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}

	// Calculate available space with overflow protection
	bavail := stat.Bavail // uint64
	bsize := stat.Bsize   // uint32

	// Check for overflow in multiplication
	if bavail > 0 && bsize > 0 && bavail > uint64(math.MaxInt64)/uint64(bsize) {
		// Handle overflow by returning MaxInt64 as available space
		return math.MaxInt64, nil
	}

	// Safe conversion after overflow check
	product := bavail * uint64(bsize) // #nosec G115 - overflow checked below
	if product > math.MaxInt64 {
		return math.MaxInt64, nil // #nosec G115 - capped at MaxInt64
	}

	// Safe conversion - already checked that product <= math.MaxInt64 above
	return int64(product), nil // #nosec G115 - overflow checked above
}
