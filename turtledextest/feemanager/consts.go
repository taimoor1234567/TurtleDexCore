package feemanager

import (
	"os"

	"github.com/turtledex/TurtleDexCore/persist"
	"github.com/turtledex/TurtleDexCore/siatest"
)

// feeManagerTestDir creates a temporary testing directory for a feemanager
// test. This should only every be called once per test. Otherwise it will
// delete the directory again.
func feeManagerTestDir(testName string) string {
	path := siatest.TestDir("feemanager", testName)
	if err := os.MkdirAll(path, persist.DefaultDiskPermissionsTest); err != nil {
		panic(err)
	}
	return path
}
