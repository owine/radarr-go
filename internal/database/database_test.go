package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindProjectRoot(t *testing.T) {
	// Change to a subdirectory to test path resolution
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create a temporary subdirectory
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir", "nested")
	err = os.MkdirAll(subDir, 0750)
	require.NoError(t, err)

	// Create a go.mod file in the temp directory
	goModPath := filepath.Join(tempDir, "go.mod")
	err = os.WriteFile(goModPath, []byte("module test\n"), 0600)
	require.NoError(t, err)

	// Change to the nested subdirectory
	err = os.Chdir(subDir)
	require.NoError(t, err)

	// Test findProjectRoot
	root, err := findProjectRoot()
	require.NoError(t, err)

	// Clean the paths for comparison (handle macOS symlinks)
	expectedPath, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)

	// Should find the directory containing go.mod
	assert.Equal(t, expectedPath, actualPath)
}

func TestFindProjectRootFromCurrentDir(t *testing.T) {
	// Test from actual project directory
	root, err := findProjectRoot()
	require.NoError(t, err)

	// Should find a go.mod file in the root
	goModPath := filepath.Join(root, "go.mod")
	_, err = os.Stat(goModPath)
	assert.NoError(t, err, "Should find go.mod in project root")

	// Should also find the migrations directory
	migrationsPath := filepath.Join(root, "migrations")
	_, err = os.Stat(migrationsPath)
	assert.NoError(t, err, "Should find migrations directory in project root")
}
