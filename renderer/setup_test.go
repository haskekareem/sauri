package renderer

import (
	"os"
	"testing"
)

// TestMain sets up any global configuration needed before running the test suite.
func TestMain(t *testing.M) {
	// Perform any setup here (e.g., logging, test env configuration)

	os.Exit(t.Run())
}
