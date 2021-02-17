package miner

import (
	"os"

	"github.com/turtledex/TurtleDexCore/persist"
	"github.com/turtledex/TurtleDexCore/siatest"
)

// minerTestDir creates a temporary testing directory for a miner test. This
// should only every be called once per test. Otherwise it will delete the
// directory again.
func minerTestDir(testName string) string {
	path := siatest.TestDir("miner", testName)
	if err := os.MkdirAll(path, persist.DefaultDiskPermissionsTest); err != nil {
		panic(err)
	}
	return path
}
