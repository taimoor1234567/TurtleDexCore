package siafile

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
	"github.com/turtledex/errors"
	"github.com/turtledex/fastrand"
	"github.com/turtledex/writeaheadlog"
)

// createLinkedBlankTurtleDexfile creates 2 TurtleDexFiles which use the same TurtleDexFile to
// store combined chunks. They reside within 'dir'.
//
//lint:file-ignore U1000 Ignore unused code, it's for future partial upload code
func createLinkedBlankTurtleDexfiles(dir string) (*TurtleDexFile, *TurtleDexFile, error) {
	// Create a wal.
	walFilePath := filepath.Join(dir, "writeaheadlog.wal")
	_, wal, err := writeaheadlog.New(walFilePath)
	if err != nil {
		return nil, nil, err
	}
	// Get parameters for the files.
	_, _, source, rc, sk, fileSize, numChunks, fileMode := newTestFileParams(1, true)
	// Create a TurtleDexFile for partial chunks.
	var partialsTurtleDexFile *TurtleDexFile
	partialsTurtleDexPath := modules.CombinedTurtleDexFilePath(rc)
	partialsTurtleDexFilePath := partialsTurtleDexPath.TurtleDexPartialsFileSysPath(dir)
	if _, err = os.Stat(partialsTurtleDexFilePath); os.IsNotExist(err) {
		partialsTurtleDexFile, err = New(partialsTurtleDexFilePath, "", wal, rc, sk, 0, fileMode, nil, false)
	} else {
		partialsTurtleDexFile, err = LoadTurtleDexFile(partialsTurtleDexFilePath, wal)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load partialsTurtleDexFile: %v", err)
	}
	/*
		 PARTIAL TODO:
			partialsEntry := &TurtleDexFileSetEntry{
				dummyEntry(partialsTurtleDexFile),
				uint64(fastrand.Intn(math.MaxInt32)),
			}
	*/
	// Create the files.
	sf1Path := filepath.Join(dir, "sf1"+modules.TurtleDexFileExtension)
	sf2Path := filepath.Join(dir, "sf2"+modules.TurtleDexFileExtension)
	sf1, err := New(sf1Path, source, wal, rc, sk, fileSize, fileMode, nil, false)
	if err != nil {
		return nil, nil, err
	}
	sf2, err := New(sf2Path, source, wal, rc, sk, fileSize, fileMode, nil, false)
	if err != nil {
		return nil, nil, err
	}
	// Check that the number of chunks in the files are correct.
	if numChunks >= 0 && sf1.numChunks != numChunks {
		return nil, nil, errors.New("createLinkedBlankTurtleDexfiles didn't create the expected number of chunks")
	}
	if numChunks >= 0 && sf2.numChunks != numChunks {
		return nil, nil, errors.New("createLinkedBlankTurtleDexfiles didn't create the expected number of chunks")
	}
	if partialsTurtleDexFile.numChunks != 0 {
		return nil, nil, errors.New("createLinkedBlankTurtleDexfiles didn't create an empty partialsTurtleDexFile")
	}
	return sf1, sf2, nil
}

// TestBackupRestoreMetadata tests that restoring a metadata from its backup
// works as expected. Especially using it as a deferred statement like we would
// use it in production code.
func TestBackupRestoreMetadata(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	sf := newTestFile()

	// Test both nil slice and regular slice.
	if fastrand.Intn(2) == 0 {
		sf.staticMetadata.Skylinks = []string{}
	} else {
		sf.staticMetadata.Skylinks = nil
	}
	if fastrand.Intn(2) == 0 {
		sf.staticMetadata.PartialChunks = []PartialChunkInfo{}
	} else {
		sf.staticMetadata.PartialChunks = nil
	}

	// Clone the metadata before modifying it.
	mdBefore := sf.staticMetadata.backup()

	// Make sure it's not the same address. Otherwise the test would later just
	// compare the pointer to itself.
	if &mdBefore == &sf.staticMetadata {
		t.Fatal("backup only copied pointer")
	}
	// To be 100% sure this works we call it like we would in the remaining
	// codebase. Deferred with a named retval.
	func() (err error) {
		// Adding this should restore the metadata later.
		defer func(backup Metadata) {
			if err != nil {
				sf.staticMetadata.restore(backup)
			}
		}(sf.staticMetadata.backup()) // NOTE: this needs to be passed in like that to work

		// Change all fields that are not static.
		sf.staticMetadata.UniqueID = TurtleDexfileUID(fmt.Sprint(fastrand.Intn(100)))
		sf.staticMetadata.FileSize = int64(fastrand.Intn(100))
		sf.staticMetadata.LocalPath = string(fastrand.Bytes(100))
		sf.staticMetadata.DisablePartialChunk = !sf.staticMetadata.DisablePartialChunk
		sf.staticMetadata.HasPartialChunk = !sf.staticMetadata.HasPartialChunk
		sf.staticMetadata.PartialChunks = nil
		if fastrand.Intn(2) == 0 { // 50% chance to be not nil
			sf.staticMetadata.PartialChunks = make([]PartialChunkInfo, fastrand.Intn(10))
		}
		sf.staticMetadata.ModTime = time.Now()
		sf.staticMetadata.ChangeTime = time.Now()
		sf.staticMetadata.AccessTime = time.Now()
		sf.staticMetadata.CreateTime = time.Now()
		sf.staticMetadata.CachedRedundancy = float64(fastrand.Intn(10))
		sf.staticMetadata.CachedUserRedundancy = float64(fastrand.Intn(10))
		sf.staticMetadata.CachedHealth = float64(fastrand.Intn(10))
		sf.staticMetadata.CachedStuckHealth = float64(fastrand.Intn(10))
		sf.staticMetadata.CachedExpiration = types.BlockHeight(fastrand.Intn(10))
		sf.staticMetadata.CachedUploadedBytes = uint64(fastrand.Intn(1000))
		sf.staticMetadata.CachedUploadProgress = float64(fastrand.Intn(100))
		sf.staticMetadata.Health = float64(fastrand.Intn(100))
		sf.staticMetadata.LastHealthCheckTime = time.Now()
		sf.staticMetadata.NumStuckChunks = fastrand.Uint64n(100)
		sf.staticMetadata.Redundancy = float64(fastrand.Intn(10))
		sf.staticMetadata.StuckHealth = float64(fastrand.Intn(100))
		sf.staticMetadata.Mode = os.FileMode(fastrand.Intn(100))
		sf.staticMetadata.UserID = int32(fastrand.Intn(100))
		sf.staticMetadata.GroupID = int32(fastrand.Intn(100))
		sf.staticMetadata.ChunkOffset = int64(fastrand.Uint64n(100))
		sf.staticMetadata.PubKeyTableOffset = int64(fastrand.Uint64n(100))
		sf.staticMetadata.Skylinks = nil
		if fastrand.Intn(2) == 0 { // 50% chance to be not nil
			sf.staticMetadata.Skylinks = make([]string, fastrand.Intn(10))
		}

		// Error occurred after changing the fields.
		return errors.New("")
	}()
	// Fields should be the same as before.
	if !reflect.DeepEqual(mdBefore, sf.staticMetadata) {
		t.Fatalf("metadata wasn't restored successfully %v %v", mdBefore, sf.staticMetadata)
	}
}
