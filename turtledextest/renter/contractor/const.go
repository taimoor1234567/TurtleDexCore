package contractor

import (
	"os"

	"github.com/turtledex/TurtleDexCore/persist"
	"github.com/turtledex/TurtleDexCore/siatest"
)

// contractorTestDir creates a temporary testing directory for a contractor
// test. This should only every be called once per test. Otherwise it will
// delete the directory again.
func contractorTestDir(testName string) string {
	path := siatest.TestDir("renter/contractor", testName)
	if err := os.MkdirAll(path, persist.DefaultDiskPermissionsTest); err != nil {
		panic(err)
	}
	return path
}
