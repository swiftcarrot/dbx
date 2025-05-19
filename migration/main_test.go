package migration

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup code before running tests

	// Run tests
	code := m.Run()

	// Cleanup after tests

	os.Exit(code)
}
