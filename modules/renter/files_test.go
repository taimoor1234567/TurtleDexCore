package renter

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turtledex/errors"
	"github.com/turtledex/fastrand"

	"github.com/turtledex/TurtleDexCore/crypto"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/modules/renter/filesystem"
	"github.com/turtledex/TurtleDexCore/persist"
)

// newTurtleDexPath returns a new TurtleDexPath for testing and panics on error
func newTurtleDexPath(str string) modules.TurtleDexPath {
	sp, err := modules.NewTurtleDexPath(str)
	if err != nil {
		panic(err)
	}
	return sp
}

// createRenterTestFile creates a test file when the test has a renter so that the
// file is properly added to the renter. It returns the TurtleDexFileSetEntry that the
// TurtleDexFile is stored in
func (r *Renter) createRenterTestFile(siaPath modules.TurtleDexPath) (*filesystem.FileNode, error) {
	// Generate erasure coder
	_, rsc := testingFileParams()
	return r.createRenterTestFileWithParams(siaPath, rsc, crypto.RandomCipherType())
}

// createRenterTestFileWithParams creates a test file when the test has a renter
// so that the file is properly added to the renter. It returns the
// TurtleDexFileSetEntry that the TurtleDexFile is stored in
func (r *Renter) createRenterTestFileWithParams(siaPath modules.TurtleDexPath, rsc modules.ErasureCoder, ct crypto.CipherType) (*filesystem.FileNode, error) {
	// create the renter/files dir if it doesn't exist
	siaFilePath := r.staticFileSystem.FilePath(siaPath)
	dir, _ := filepath.Split(siaFilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	// Create File
	up := modules.FileUploadParams{
		Source:      "",
		TurtleDexPath:     siaPath,
		ErasureCode: rsc,
	}
	err := r.staticFileSystem.NewTurtleDexFile(up.TurtleDexPath, up.Source, up.ErasureCode, crypto.GenerateTurtleDexKey(ct), 1000, persist.DefaultDiskPermissionsTest, false)
	if err != nil {
		return nil, err
	}
	return r.staticFileSystem.OpenTurtleDexFile(up.TurtleDexPath)
}

// newRenterTestFile creates a test file when the test has a renter so that the
// file is properly added to the renter. It returns the TurtleDexFileSetEntry that the
// TurtleDexFile is stored in
func (r *Renter) newRenterTestFile() (*filesystem.FileNode, error) {
	// Generate name and erasure coding
	siaPath, rsc := testingFileParams()
	return r.createRenterTestFileWithParams(siaPath, rsc, crypto.RandomCipherType())
}

// TestRenterFileListLocalPath verifies that FileList() returns the correct
// local path information for an uploaded file.
func TestRenterFileListLocalPath(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	rt, err := newRenterTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	id := rt.renter.mu.Lock()
	entry, _ := rt.renter.newRenterTestFile()
	if err := entry.SetLocalPath("TestPath"); err != nil {
		t.Fatal(err)
	}
	rt.renter.mu.Unlock(id)
	files, err := rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatal("wrong number of files, got", len(files), "wanted one")
	}
	if files[0].LocalPath != "TestPath" {
		t.Fatal("file had wrong LocalPath: got", files[0].LocalPath, "wanted TestPath")
	}
}

// TestRenterDeleteFile probes the DeleteFile method of the renter type.
func TestRenterDeleteFile(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	rt, err := newRenterTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Delete a file from an empty renter.
	siaPath, err := modules.NewTurtleDexPath("dne")
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.DeleteFile(siaPath)
	// NOTE: using strings.Contains because errors.Contains does not recognize
	// errors when errors.Extend is used
	if !strings.Contains(err.Error(), filesystem.ErrNotExist.Error()) {
		t.Errorf("Expected error to contain %v but got '%v'", filesystem.ErrNotExist, err)
	}

	// Put a file in the renter.
	entry, err := rt.renter.newRenterTestFile()
	if err != nil {
		t.Fatal(err)
	}
	// Delete a different file.
	siaPathOne, err := modules.NewTurtleDexPath("one")
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.DeleteFile(siaPathOne)
	// NOTE: using strings.Contains because errors.Contains does not recognize
	// errors when errors.Extend is used
	if !strings.Contains(err.Error(), filesystem.ErrNotExist.Error()) {
		t.Errorf("Expected error to contain %v but got '%v'", filesystem.ErrNotExist, err)
	}
	// Delete the file.
	siapath := rt.renter.staticFileSystem.FileTurtleDexPath(entry)

	if err := entry.Close(); err != nil {
		t.Fatal(err)
	}
	err = rt.renter.DeleteFile(siapath)
	if err != nil {
		t.Fatal(err)
	}
	files, err := rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Error("file was deleted, but is still reported in FileList")
	}
	// Confirm that file was removed from TurtleDexFileSet
	_, err = rt.renter.staticFileSystem.OpenTurtleDexFile(siapath)
	if err == nil {
		t.Fatal("Deleted file still found in staticFileSet")
	}

	// Put a file in the renter, then rename it.
	entry2, err := rt.renter.newRenterTestFile()
	if err != nil {
		t.Fatal(err)
	}
	siaPath1, err := modules.NewTurtleDexPath("1")
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(rt.renter.staticFileSystem.FileTurtleDexPath(entry2), siaPath1) // set name to "1"
	if err != nil {
		t.Fatal(err)
	}
	siapath2 := rt.renter.staticFileSystem.FileTurtleDexPath(entry2)
	entry2.Close()
	siapath2 = rt.renter.staticFileSystem.FileTurtleDexPath(entry2)
	err = rt.renter.RenameFile(siapath2, siaPathOne)
	if err != nil {
		t.Fatal(err)
	}
	// Call delete on the previous name.
	err = rt.renter.DeleteFile(siaPath1)
	// NOTE: using strings.Contains because errors.Contains does not recognize
	// errors when errors.Extend is used
	if !strings.Contains(err.Error(), filesystem.ErrNotExist.Error()) {
		t.Errorf("Expected error to contain %v but got '%v'", filesystem.ErrNotExist, err)
	}
	// Call delete on the new name.
	err = rt.renter.DeleteFile(siaPathOne)
	if err != nil {
		t.Error(err)
	}

	// Check that all .sia files have been deleted.
	var walkStr string
	rt.renter.staticFileSystem.Walk(modules.RootTurtleDexPath(), func(path string, _ os.FileInfo, _ error) error {
		// capture only .sia files
		if filepath.Ext(path) == ".sia" {
			rel, _ := filepath.Rel(rt.renter.staticFileSystem.Root(), path) // strip testdir prefix
			walkStr += rel
		}
		return nil
	})
	expWalkStr := ""
	if walkStr != expWalkStr {
		t.Fatalf("Bad walk string: expected %q, got %q", expWalkStr, walkStr)
	}
}

// TestRenterDeleteFileMissingParent tries to delete a file for which the parent
// has been deleted before.
func TestRenterDeleteFileMissingParent(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	rt, err := newRenterTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Put a file in the renter.
	siaPath, err := modules.NewTurtleDexPath("parent/file")
	if err != nil {
		t.Fatal(err)
	}
	dirTurtleDexPath, err := siaPath.Dir()
	if err != nil {
		t.Fatal(err)
	}
	siaPath, rsc := testingFileParams()
	up := modules.FileUploadParams{
		Source:      "",
		TurtleDexPath:     siaPath,
		ErasureCode: rsc,
	}
	err = rt.renter.staticFileSystem.NewTurtleDexFile(up.TurtleDexPath, up.Source, up.ErasureCode, crypto.GenerateTurtleDexKey(crypto.RandomCipherType()), 1000, persist.DefaultDiskPermissionsTest, false)
	if err != nil {
		t.Fatal(err)
	}
	// Delete the parent.
	err = rt.renter.staticFileSystem.DeleteFile(dirTurtleDexPath)
	// NOTE: using strings.Contains because errors.Contains does not recognize
	// errors when errors.Extend is used
	if !strings.Contains(err.Error(), filesystem.ErrNotExist.Error()) {
		t.Errorf("Expected error to contain %v but got '%v'", filesystem.ErrNotExist, err)
	}
	// Delete the file. This should not return an error since it's already
	// deleted implicitly.
	if err := rt.renter.staticFileSystem.DeleteFile(up.TurtleDexPath); err != nil {
		t.Fatal(err)
	}
}

// TestRenterFileList probes the FileList method of the renter type.
func TestRenterFileList(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	rt, err := newRenterTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Get the file list of an empty renter.
	files, err := rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatal("FileList has non-zero length for empty renter?")
	}

	// Put a file in the renter.
	entry1, _ := rt.renter.newRenterTestFile()
	files, err = rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatal("FileList is not returning the only file in the renter")
	}
	entry1SP := rt.renter.staticFileSystem.FileTurtleDexPath(entry1)
	if !files[0].TurtleDexPath.Equals(entry1SP) {
		t.Error("FileList is not returning the correct filename for the only file")
	}

	// Put multiple files in the renter.
	entry2, _ := rt.renter.newRenterTestFile()
	entry2SP := rt.renter.staticFileSystem.FileTurtleDexPath(entry2)
	files, err = rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("Expected %v files, got %v", 2, len(files))
	}
	files, err = rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !((files[0].TurtleDexPath.Equals(entry1SP) || files[0].TurtleDexPath.Equals(entry2SP)) &&
		(files[1].TurtleDexPath.Equals(entry1SP) || files[1].TurtleDexPath.Equals(entry2SP)) &&
		(files[0].TurtleDexPath != files[1].TurtleDexPath)) {
		t.Log("files[0].TurtleDexPath", files[0].TurtleDexPath)
		t.Log("files[1].TurtleDexPath", files[1].TurtleDexPath)
		t.Log("file1.TurtleDexPath()", rt.renter.staticFileSystem.FileTurtleDexPath(entry1).String())
		t.Log("file2.TurtleDexPath()", rt.renter.staticFileSystem.FileTurtleDexPath(entry2).String())
		t.Error("FileList is returning wrong names for the files")
	}
}

// TestRenterRenameFile probes the rename method of the renter.
func TestRenterRenameFile(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	rt, err := newRenterTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Rename a file that doesn't exist.
	siaPath1, err := modules.NewTurtleDexPath("1")
	if err != nil {
		t.Fatal(err)
	}
	siaPath1a, err := modules.NewTurtleDexPath("1a")
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(siaPath1, siaPath1a)
	if err.Error() != filesystem.ErrNotExist.Error() {
		t.Errorf("Expected '%v' got '%v'", filesystem.ErrNotExist, err)
	}

	// Get the filesystem.
	sfs := rt.renter.staticFileSystem

	// Rename a file that does exist.
	entry, _ := rt.renter.newRenterTestFile()
	var sp modules.TurtleDexPath
	if err := sp.FromSysPath(entry.TurtleDexFilePath(), sfs.DirPath(modules.RootTurtleDexPath())); err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(sp, siaPath1)
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(siaPath1, siaPath1a)
	if err != nil {
		t.Fatal(err)
	}
	files, err := rt.renter.FileListCollect(modules.RootTurtleDexPath(), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatal("FileList has unexpected number of files:", len(files))
	}
	if !files[0].TurtleDexPath.Equals(siaPath1a) {
		t.Errorf("RenameFile failed: expected %v, got %v", siaPath1a.String(), files[0].TurtleDexPath)
	}
	// Confirm TurtleDexFileSet was updated
	_, err = rt.renter.staticFileSystem.OpenTurtleDexFile(siaPath1a)
	if err != nil {
		t.Fatal("renter staticFileSet not updated to new file name:", err)
	}
	_, err = rt.renter.staticFileSystem.OpenTurtleDexFile(siaPath1)
	if err == nil {
		t.Fatal("old name not removed from renter staticFileSet")
	}
	// Rename a file to an existing name.
	entry2, err := rt.renter.newRenterTestFile()
	if err != nil {
		t.Fatal(err)
	}
	var sp2 modules.TurtleDexPath
	if err := sp2.FromSysPath(entry2.TurtleDexFilePath(), sfs.DirPath(modules.RootTurtleDexPath())); err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(sp2, siaPath1) // Rename to "1"
	if err != nil {
		t.Fatal(err)
	}
	entry2.Close()
	err = rt.renter.RenameFile(siaPath1, siaPath1a)
	if !errors.Contains(err, filesystem.ErrExists) {
		t.Fatal("Expecting ErrExists, got", err)
	}
	// Rename a file to the same name.
	err = rt.renter.RenameFile(siaPath1, siaPath1)
	if !errors.Contains(err, filesystem.ErrExists) {
		t.Fatal("Expecting ErrExists, got", err)
	}

	// Confirm ability to rename file
	siaPath1b, err := modules.NewTurtleDexPath("1b")
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(siaPath1, siaPath1b)
	if err != nil {
		t.Fatal(err)
	}
	// Rename file that would create a directory
	siaPathWithDir, err := modules.NewTurtleDexPath("new/name/with/dir/test")
	if err != nil {
		t.Fatal(err)
	}
	err = rt.renter.RenameFile(siaPath1b, siaPathWithDir)
	if err != nil {
		t.Fatal(err)
	}

	// Confirm directory metadatas exist
	for !siaPathWithDir.Equals(modules.RootTurtleDexPath()) {
		siaPathWithDir, err = siaPathWithDir.Dir()
		if err != nil {
			t.Fatal(err)
		}
		_, err = rt.renter.staticFileSystem.OpenTurtleDexDir(siaPathWithDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestRenterFileDir tests that the renter files are uploaded to the files
// directory and not the root directory of the renter.
func TestRenterFileDir(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	rt, err := newRenterTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Create local file to upload
	localDir := filepath.Join(rt.dir, "files")
	if err := os.MkdirAll(localDir, 0700); err != nil {
		t.Fatal(err)
	}
	size := 100
	fileName := fmt.Sprintf("%dbytes %s", size, hex.EncodeToString(fastrand.Bytes(4)))
	source := filepath.Join(localDir, fileName)
	bytes := fastrand.Bytes(size)
	if err := ioutil.WriteFile(source, bytes, 0600); err != nil {
		t.Fatal(err)
	}

	// Upload local file
	ec := modules.NewRSCodeDefault()
	siaPath, err := modules.NewTurtleDexPath(fileName)
	if err != nil {
		t.Fatal(err)
	}
	params := modules.FileUploadParams{
		Source:      source,
		TurtleDexPath:     siaPath,
		ErasureCode: ec,
	}
	err = rt.renter.Upload(params)
	if err != nil {
		t.Fatal("failed to upload file:", err)
	}

	// Get file and check siapath
	f, err := rt.renter.File(siaPath)
	if err != nil {
		t.Fatal(err)
	}
	if !f.TurtleDexPath.Equals(siaPath) {
		t.Fatalf("siapath not set as expected: got %v expected %v", f.TurtleDexPath, fileName)
	}

	// Confirm .sia file exists on disk in the TurtleDexpathRoot directory
	renterDir := filepath.Join(rt.dir, modules.RenterDir)
	siapathRootDir := filepath.Join(renterDir, modules.FileSystemRoot)
	fullPath := siaPath.TurtleDexFileSysPath(siapathRootDir)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatal("No .sia file found on disk")
	}
}
