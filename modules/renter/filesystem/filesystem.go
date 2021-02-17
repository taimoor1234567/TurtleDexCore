package filesystem

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/turtledex/TurtleDexCore/build"
	"github.com/turtledex/TurtleDexCore/crypto"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/modules/renter/filesystem/ttdxdir"
	"github.com/turtledex/TurtleDexCore/modules/renter/filesystem/siafile"
	"github.com/turtledex/TurtleDexCore/persist"
	"github.com/turtledex/errors"
	"github.com/turtledex/fastrand"
	"github.com/turtledex/writeaheadlog"
)

var (
	// ErrNotExist is returned when a file or folder can't be found on disk.
	ErrNotExist = errors.New("path does not exist")

	// ErrExists is returned when a file or folder already exists at a given
	// location.
	ErrExists = errors.New("a file or folder already exists at the specified path")

	// ErrDeleteFileIsDir is returned when the file delete method is used but
	// the filename corresponds to a directory
	ErrDeleteFileIsDir = errors.New("cannot delete file, file is a directory")
)

type (
	// FileSystem implements a thread-safe filesystem for TurtleDex for loading
	// TurtleDexFiles, TurtleDexDirs and potentially other supported TurtleDex types in the
	// future.
	FileSystem struct {
		DirNode
	}

	// node is a struct that contains the common fields of every node.
	node struct {
		// fields that all copies of a node share.
		path      *string
		parent    *DirNode
		name      *string
		staticWal *writeaheadlog.WAL
		threads   map[threadUID]struct{} // tracks all the threadUIDs of evey copy of the node
		staticLog *persist.Logger
		staticUID uint64
		mu        *sync.Mutex

		// fields that differ between copies of the same node.
		threadUID threadUID // unique ID of a copy of a node
	}
	threadUID uint64
)

// newNode is a convenience function to initialize a node.
func newNode(parent *DirNode, path, name string, uid threadUID, wal *writeaheadlog.WAL, log *persist.Logger) node {
	return node{
		path:      &path,
		parent:    parent,
		name:      &name,
		staticLog: log,
		staticUID: newInode(),
		staticWal: wal,
		threads:   make(map[threadUID]struct{}),
		threadUID: uid,
		mu:        new(sync.Mutex),
	}
}

// managedLockWithParent is a helper method which correctly acquires the lock of
// a node and it's parent. If no parent it available it will return 'nil'. In
// either case the node and potential parent will be locked after the call.
func (n *node) managedLockWithParent() *DirNode {
	var parent *DirNode
	for {
		// If a parent exists, we need to lock it while closing a child.
		n.mu.Lock()
		parent = n.parent
		n.mu.Unlock()
		if parent != nil {
			parent.mu.Lock()
		}
		n.mu.Lock()
		if n.parent != parent {
			n.mu.Unlock()
			parent.mu.Unlock()
			continue // try again
		}
		break
	}
	return parent
}

// NID returns the node's unique identifier.
func (n *node) Inode() uint64 {
	return n.staticUID
}

// newThreadUID returns a random threadUID to be used as the threadUID in the
// threads map of the node.
func newThreadUID() threadUID {
	return threadUID(fastrand.Uint64n(math.MaxUint64))
}

// newInode will create a unique identifier for a filesystem node.
//
// TODO: replace this with a function that doesn't repeat itself.
func newInode() uint64 {
	return fastrand.Uint64n(math.MaxUint64)
}

// nodeTurtleDexPath returns the TurtleDexPath of a node relative to a root path.
func nodeTurtleDexPath(rootPath string, n *node) (sp modules.TurtleDexPath) {
	if err := sp.FromSysPath(n.managedAbsPath(), rootPath); err != nil {
		build.Critical("FileSystem.managedTurtleDexPath: should never fail", err)
	}
	return sp
}

// closeNode removes a thread from the node's threads map. This should only be
// called from within other 'close' methods.
func (n *node) closeNode() {
	if _, exists := n.threads[n.threadUID]; !exists {
		build.Critical("threaduid doesn't exist in threads map: ", n.threadUID, len(n.threads))
	}
	delete(n.threads, n.threadUID)
}

// absPath returns the absolute path of the node.
func (n *node) absPath() string {
	return *n.path
}

// managedAbsPath returns the absolute path of the node.
func (n *node) managedAbsPath() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.absPath()
}

// New creates a new FileSystem at the specified root path. The folder will be
// created if it doesn't exist already.
func New(root string, log *persist.Logger, wal *writeaheadlog.WAL) (*FileSystem, error) {
	fs := &FileSystem{
		DirNode: DirNode{
			// The root doesn't require a parent, a name or uid.
			node:        newNode(nil, root, "", 0, wal, log),
			directories: make(map[string]*DirNode),
			files:       make(map[string]*FileNode),
			lazyTurtleDexDir:  new(*ttdxdir.TurtleDexDir),
		},
	}
	// Prepare root folder.
	err := fs.NewTurtleDexDir(modules.RootTurtleDexPath(), modules.DefaultDirPerm)
	if err != nil && !errors.Contains(err, ErrExists) {
		return nil, err
	}
	return fs, nil
}

// AddTurtleDexFileFromReader adds an existing TurtleDexFile to the set and stores it on
// disk. If the exact same file already exists, this is a no-op. If a file
// already exists with a different UID, the UID will be updated and a unique
// path will be chosen. If no file exists, the UID will be updated but the path
// remains the same.
func (fs *FileSystem) AddTurtleDexFileFromReader(rs io.ReadSeeker, siaPath modules.TurtleDexPath) (err error) {
	// Load the file.
	path := fs.FilePath(siaPath)
	sf, chunks, err := siafile.LoadTurtleDexFileFromReaderWithChunks(rs, path, fs.staticWal)
	if err != nil {
		return err
	}
	// Create dir with same Mode as file if it doesn't exist already and open
	// it.
	dirTurtleDexPath, err := siaPath.Dir()
	if err != nil {
		return err
	}
	if err := fs.managedNewTurtleDexDir(dirTurtleDexPath, sf.Mode()); err != nil {
		return err
	}
	dir, err := fs.managedOpenDir(dirTurtleDexPath.String())
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Compose(err, dir.Close())
	}()
	// Add the file to the dir.
	return dir.managedNewTurtleDexFileFromExisting(sf, chunks)
}

// CachedFileInfo returns the cached File Information of the siafile
func (fs *FileSystem) CachedFileInfo(siaPath modules.TurtleDexPath) (modules.FileInfo, error) {
	return fs.managedFileInfo(siaPath, true, nil, nil, nil)
}

// CachedList lists the files and directories within a TurtleDexDir.
func (fs *FileSystem) CachedList(siaPath modules.TurtleDexPath, recursive bool, flf modules.FileListFunc, dlf modules.DirListFunc) error {
	return fs.managedList(siaPath, recursive, true, nil, nil, nil, flf, dlf)
}

// CachedListOnNode will return the files and directories within a given ttdxdir
// node in a non-recursive way.
func (fs *FileSystem) CachedListOnNode(d *DirNode) (fis []modules.FileInfo, dis []modules.DirectoryInfo, err error) {
	var fmu, dmu sync.Mutex
	flf := func(fi modules.FileInfo) {
		fmu.Lock()
		fis = append(fis, fi)
		fmu.Unlock()
	}
	dlf := func(di modules.DirectoryInfo) {
		dmu.Lock()
		dis = append(dis, di)
		dmu.Unlock()
	}
	err = d.managedList(fs.managedAbsPath(), false, true, nil, nil, nil, flf, dlf)

	// Sort slices by TurtleDexPath.
	sort.Slice(dis, func(i, j int) bool {
		return dis[i].TurtleDexPath.String() < dis[j].TurtleDexPath.String()
	})
	sort.Slice(fis, func(i, j int) bool {
		return fis[i].TurtleDexPath.String() < fis[j].TurtleDexPath.String()
	})
	return
}

// DeleteDir deletes a dir from the filesystem. The dir will be marked as
// 'deleted' which should cause all remaining instances of the dir to be close
// shortly. Only when all instances of the dir are closed it will be removed
// from the tree. This means that as long as the deletion is in progress, no new
// file of the same path can be created and the existing file can't be opened
// until all instances of it are closed.
func (fs *FileSystem) DeleteDir(siaPath modules.TurtleDexPath) error {
	return fs.managedDeleteDir(siaPath.String())
}

// DeleteFile deletes a file from the filesystem. The file will be marked as
// 'deleted' which should cause all remaining instances of the file to be closed
// shortly. Only when all instances of the file are closed it will be removed
// from the tree. This means that as long as the deletion is in progress, no new
// file of the same path can be created and the existing file can't be opened
// until all instances of it are closed.
func (fs *FileSystem) DeleteFile(siaPath modules.TurtleDexPath) error {
	return fs.managedDeleteFile(siaPath.String())
}

// DirInfo returns the Directory Information of the ttdxdir
func (fs *FileSystem) DirInfo(siaPath modules.TurtleDexPath) (_ modules.DirectoryInfo, err error) {
	dir, err := fs.managedOpenDir(siaPath.String())
	if err != nil {
		return modules.DirectoryInfo{}, nil
	}
	defer func() {
		err = errors.Compose(err, dir.Close())
	}()
	di, err := dir.managedInfo(siaPath)
	if err != nil {
		return modules.DirectoryInfo{}, err
	}
	di.TurtleDexPath = siaPath
	return di, nil
}

// DirNodeInfo will return the DirectoryInfo of a ttdxdir given the node. This is
// more efficient than calling fs.DirInfo.
func (fs *FileSystem) DirNodeInfo(n *DirNode) (modules.DirectoryInfo, error) {
	sp := fs.DirTurtleDexPath(n)
	return n.managedInfo(sp)
}

// FileInfo returns the File Information of the siafile
func (fs *FileSystem) FileInfo(siaPath modules.TurtleDexPath, offline map[string]bool, goodForRenew map[string]bool, contracts map[string]modules.RenterContract) (modules.FileInfo, error) {
	return fs.managedFileInfo(siaPath, false, offline, goodForRenew, contracts)
}

// FileNodeInfo returns the FileInfo of a siafile given the node for the
// siafile. This is faster than calling fs.FileInfo.
func (fs *FileSystem) FileNodeInfo(n *FileNode) (modules.FileInfo, error) {
	sp := fs.FileTurtleDexPath(n)
	return n.staticCachedInfo(sp)
}

// List lists the files and directories within a TurtleDexDir.
func (fs *FileSystem) List(siaPath modules.TurtleDexPath, recursive bool, offlineMap, goodForRenewMap map[string]bool, contractsMap map[string]modules.RenterContract, flf modules.FileListFunc, dlf modules.DirListFunc) error {
	return fs.managedList(siaPath, recursive, false, offlineMap, goodForRenewMap, contractsMap, flf, dlf)
}

// FileExists checks to see if a file with the provided siaPath already exists
// in the renter.
func (fs *FileSystem) FileExists(siaPath modules.TurtleDexPath) (bool, error) {
	path := fs.FilePath(siaPath)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// FilePath converts a TurtleDexPath into a file's system path.
func (fs *FileSystem) FilePath(siaPath modules.TurtleDexPath) string {
	return siaPath.TurtleDexFileSysPath(fs.managedAbsPath())
}

// NewTurtleDexDir creates the folder for the specified siaPath.
func (fs *FileSystem) NewTurtleDexDir(siaPath modules.TurtleDexPath, mode os.FileMode) error {
	return fs.managedNewTurtleDexDir(siaPath, mode)
}

// NewTurtleDexFile creates a TurtleDexFile at the specified siaPath.
func (fs *FileSystem) NewTurtleDexFile(siaPath modules.TurtleDexPath, source string, ec modules.ErasureCoder, mk crypto.CipherKey, fileSize uint64, fileMode os.FileMode, disablePartialUpload bool) error {
	// Create TurtleDexDir for file.
	dirTurtleDexPath, err := siaPath.Dir()
	if err != nil {
		return err
	}
	if err = fs.NewTurtleDexDir(dirTurtleDexPath, fileMode); err != nil {
		return errors.AddContext(err, fmt.Sprintf("failed to create TurtleDexDir %v for TurtleDexFile %v", dirTurtleDexPath.String(), siaPath.String()))
	}
	return fs.managedNewTurtleDexFile(siaPath.String(), source, ec, mk, fileSize, fileMode, disablePartialUpload)
}

// ReadDir reads all the fileinfos of the specified dir.
func (fs *FileSystem) ReadDir(siaPath modules.TurtleDexPath) ([]os.FileInfo, error) {
	// Open dir.
	dirPath := siaPath.TurtleDexDirSysPath(fs.managedAbsPath())
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	// Read it and close it.
	fis, err1 := f.Readdir(-1)
	err2 := f.Close()
	err = errors.Compose(err1, err2)
	return fis, err
}

// DirExists checks to see if a dir with the provided siaPath already exists in
// the renter.
func (fs *FileSystem) DirExists(siaPath modules.TurtleDexPath) (bool, error) {
	path := fs.DirPath(siaPath)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// DirPath converts a TurtleDexPath into a dir's system path.
func (fs *FileSystem) DirPath(siaPath modules.TurtleDexPath) string {
	return siaPath.TurtleDexDirSysPath(fs.managedAbsPath())
}

// Root returns the root system path of the FileSystem.
func (fs *FileSystem) Root() string {
	return fs.DirPath(modules.RootTurtleDexPath())
}

// FileTurtleDexPath returns the TurtleDexPath of a file node.
func (fs *FileSystem) FileTurtleDexPath(n *FileNode) (sp modules.TurtleDexPath) {
	return fs.managedTurtleDexPath(&n.node)
}

// DirTurtleDexPath returns the TurtleDexPath of a dir node.
func (fs *FileSystem) DirTurtleDexPath(n *DirNode) (sp modules.TurtleDexPath) {
	return fs.managedTurtleDexPath(&n.node)
}

// UpdateDirMetadata updates the metadata of a TurtleDexDir.
func (fs *FileSystem) UpdateDirMetadata(siaPath modules.TurtleDexPath, metadata ttdxdir.Metadata) (err error) {
	dir, err := fs.OpenTurtleDexDir(siaPath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Compose(err, dir.Close())
	}()
	return dir.UpdateMetadata(metadata)
}

// managedTurtleDexPath returns the TurtleDexPath of a node.
func (fs *FileSystem) managedTurtleDexPath(n *node) modules.TurtleDexPath {
	return nodeTurtleDexPath(fs.managedAbsPath(), n)
}

// Stat is a wrapper for os.Stat which takes a TurtleDexPath as an argument instead of
// a system path.
func (fs *FileSystem) Stat(siaPath modules.TurtleDexPath) (os.FileInfo, error) {
	path := siaPath.TurtleDexDirSysPath(fs.managedAbsPath())
	return os.Stat(path)
}

// Walk is a wrapper for filepath.Walk which takes a TurtleDexPath as an argument
// instead of a system path.
func (fs *FileSystem) Walk(siaPath modules.TurtleDexPath, walkFn filepath.WalkFunc) error {
	dirPath := siaPath.TurtleDexDirSysPath(fs.managedAbsPath())
	return filepath.Walk(dirPath, walkFn)
}

// WriteFile is a wrapper for ioutil.WriteFile which takes a TurtleDexPath as an
// argument instead of a system path.
func (fs *FileSystem) WriteFile(siaPath modules.TurtleDexPath, data []byte, perm os.FileMode) error {
	path := siaPath.TurtleDexFileSysPath(fs.managedAbsPath())
	return ioutil.WriteFile(path, data, perm)
}

// NewTurtleDexFileFromLegacyData creates a new TurtleDexFile from data that was previously loaded
// from a legacy file.
func (fs *FileSystem) NewTurtleDexFileFromLegacyData(fd siafile.FileData) (_ *FileNode, err error) {
	// Get file's TurtleDexPath.
	sp, err := modules.UserFolder.Join(fd.Name)
	if err != nil {
		return nil, err
	}
	// Get siapath of dir.
	dirTurtleDexPath, err := sp.Dir()
	if err != nil {
		return nil, err
	}
	// Create the dir if it doesn't exist.
	if err := fs.NewTurtleDexDir(dirTurtleDexPath, 0755); err != nil {
		return nil, err
	}
	// Open dir.
	dir, err := fs.managedOpenDir(dirTurtleDexPath.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Compose(err, dir.Close())
	}()
	// Add the file to the dir.
	return dir.managedNewTurtleDexFileFromLegacyData(sp.Name(), fd)
}

// OpenTurtleDexDir opens a TurtleDexDir and adds it and all of its parents to the
// filesystem tree.
func (fs *FileSystem) OpenTurtleDexDir(siaPath modules.TurtleDexPath) (*DirNode, error) {
	return fs.OpenTurtleDexDirCustom(siaPath, false)
}

// OpenTurtleDexDirCustom opens a TurtleDexDir and adds it and all of its parents to the
// filesystem tree. If create is true it will create the dir if it doesn't
// exist.
func (fs *FileSystem) OpenTurtleDexDirCustom(siaPath modules.TurtleDexPath, create bool) (*DirNode, error) {
	dn, err := fs.managedOpenTurtleDexDir(siaPath)
	if create && errors.Contains(err, ErrNotExist) {
		// If ttdxdir doesn't exist create one
		err = fs.NewTurtleDexDir(siaPath, modules.DefaultDirPerm)
		if err != nil && !errors.Contains(err, ErrExists) {
			return nil, err
		}
		return fs.managedOpenTurtleDexDir(siaPath)
	}
	return dn, err
}

// OpenTurtleDexFile opens a TurtleDexFile and adds it and all of its parents to the
// filesystem tree.
func (fs *FileSystem) OpenTurtleDexFile(siaPath modules.TurtleDexPath) (*FileNode, error) {
	sf, err := fs.managedOpenFile(siaPath.String())
	if err != nil {
		return nil, err
	}
	return sf, nil
}

// RenameFile renames the file with oldTurtleDexPath to newTurtleDexPath.
func (fs *FileSystem) RenameFile(oldTurtleDexPath, newTurtleDexPath modules.TurtleDexPath) (err error) {
	// Open TurtleDexDir for file at old location.
	oldDirTurtleDexPath, err := oldTurtleDexPath.Dir()
	if err != nil {
		return err
	}
	oldDir, err := fs.managedOpenTurtleDexDir(oldDirTurtleDexPath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Compose(err, oldDir.Close())
	}()
	// Open the file.
	sf, err := oldDir.managedOpenFile(oldTurtleDexPath.Name())
	if errors.Contains(err, ErrNotExist) {
		return ErrNotExist
	}
	if err != nil {
		return errors.AddContext(err, "failed to open file for renaming")
	}
	defer func() {
		err = errors.Compose(err, sf.Close())
	}()

	// Create and Open TurtleDexDir for file at new location.
	newDirTurtleDexPath, err := newTurtleDexPath.Dir()
	if err != nil {
		return err
	}
	if err := fs.NewTurtleDexDir(newDirTurtleDexPath, sf.managedMode()); err != nil {
		return errors.AddContext(err, fmt.Sprintf("failed to create TurtleDexDir %v for TurtleDexFile %v", newDirTurtleDexPath.String(), oldTurtleDexPath.String()))
	}
	newDir, err := fs.managedOpenTurtleDexDir(newDirTurtleDexPath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Compose(err, newDir.Close())
	}()
	// Rename the file.
	return sf.managedRename(newTurtleDexPath.Name(), oldDir, newDir)
}

// RenameDir takes an existing directory and changes the path. The original
// directory must exist, and there must not be any directory that already has
// the replacement path.  All sia files within directory will also be renamed
func (fs *FileSystem) RenameDir(oldTurtleDexPath, newTurtleDexPath modules.TurtleDexPath) error {
	// Open TurtleDexDir for parent dir at old location.
	oldDirTurtleDexPath, err := oldTurtleDexPath.Dir()
	if err != nil {
		return err
	}
	oldDir, err := fs.managedOpenTurtleDexDir(oldDirTurtleDexPath)
	if err != nil {
		return err
	}
	defer func() {
		oldDir.Close()
	}()
	// Open the dir to rename.
	sd, err := oldDir.managedOpenDir(oldTurtleDexPath.Name())
	if errors.Contains(err, ErrNotExist) {
		return ErrNotExist
	}
	if err != nil {
		return errors.AddContext(err, "failed to open file for renaming")
	}
	defer func() {
		sd.Close()
	}()

	// Create and Open parent TurtleDexDir for dir at new location.
	newDirTurtleDexPath, err := newTurtleDexPath.Dir()
	if err != nil {
		return err
	}
	md, err := sd.Metadata()
	if err != nil {
		return err
	}
	if err := fs.NewTurtleDexDir(newDirTurtleDexPath, md.Mode); err != nil {
		return errors.AddContext(err, fmt.Sprintf("failed to create TurtleDexDir %v for TurtleDexFile %v", newDirTurtleDexPath.String(), oldTurtleDexPath.String()))
	}
	newDir, err := fs.managedOpenTurtleDexDir(newDirTurtleDexPath)
	if err != nil {
		return err
	}
	defer func() {
		newDir.Close()
	}()
	// Rename the dir.
	err = sd.managedRename(newTurtleDexPath.Name(), oldDir, newDir)
	return err
}

// managedDeleteFile opens the parent folder of the file to delete and calls
// managedDeleteFile on it.
func (fs *FileSystem) managedDeleteFile(relPath string) (err error) {
	// Open the folder that contains the file.
	dirPath, fileName := filepath.Split(relPath)
	var dir *DirNode
	if dirPath == string(filepath.Separator) || dirPath == "." || dirPath == "" {
		dir = &fs.DirNode // file is in the root dir
	} else {
		var err error
		dir, err = fs.managedOpenDir(filepath.Dir(relPath))
		if err != nil {
			return errors.AddContext(err, "failed to open parent dir of file")
		}
		// Close the dir since we are not returning it. The open file keeps it
		// loaded in memory.
		defer func() {
			err = errors.Compose(err, dir.Close())
		}()
	}
	return dir.managedDeleteFile(fileName)
}

// managedDeleteDir opens the parent folder of the dir to delete and calls
// managedDelete on it.
func (fs *FileSystem) managedDeleteDir(path string) (err error) {
	// Open the dir.
	dir, err := fs.managedOpenDir(path)
	if err != nil {
		return errors.AddContext(err, "failed to open parent dir of file")
	}
	// Close the dir since we are not returning it. The open file keeps it
	// loaded in memory.
	defer func() {
		err = errors.Compose(err, dir.Close())
	}()
	return dir.managedDelete()
}

// managedFileInfo returns the FileInfo of the siafile.
func (fs *FileSystem) managedFileInfo(siaPath modules.TurtleDexPath, cached bool, offline map[string]bool, goodForRenew map[string]bool, contracts map[string]modules.RenterContract) (_ modules.FileInfo, err error) {
	// Open the file.
	file, err := fs.managedOpenFile(siaPath.String())
	if err != nil {
		return modules.FileInfo{}, err
	}
	defer func() {
		err = errors.Compose(err, file.Close())
	}()
	if cached {
		return file.staticCachedInfo(siaPath)
	}
	return file.managedFileInfo(siaPath, offline, goodForRenew, contracts)
}

// managedList returns the files and dirs within the TurtleDexDir specified by siaPath.
// offlineMap, goodForRenewMap and contractMap don't need to be provided if
// 'cached' is set to 'true'.
func (fs *FileSystem) managedList(siaPath modules.TurtleDexPath, recursive, cached bool, offlineMap map[string]bool, goodForRenewMap map[string]bool, contractsMap map[string]modules.RenterContract, flf modules.FileListFunc, dlf modules.DirListFunc) (err error) {
	// Open the folder.
	dir, err := fs.managedOpenDir(siaPath.String())
	if err != nil {
		return errors.AddContext(err, fmt.Sprintf("failed to open folder '%v' specified by FileList", siaPath))
	}
	defer func() {
		err = errors.Compose(err, dir.Close())
	}()
	return dir.managedList(fs.managedAbsPath(), recursive, cached, offlineMap, goodForRenewMap, contractsMap, flf, dlf)
}

// managedNewTurtleDexDir creates the folder at the specified siaPath.
func (fs *FileSystem) managedNewTurtleDexDir(siaPath modules.TurtleDexPath, mode os.FileMode) (err error) {
	// If siaPath is the root dir we just create the metadata for it.
	if siaPath.IsRoot() {
		fs.mu.Lock()
		defer fs.mu.Unlock()
		dirPath := siaPath.TurtleDexDirSysPath(fs.absPath())
		_, err := ttdxdir.New(dirPath, fs.absPath(), mode, fs.staticWal)
		// If the TurtleDexDir already exists on disk, return without an error.
		if os.IsExist(err) {
			return nil // nothing to do
		}
		return err
	}
	// If siaPath isn't the root dir we need to grab the parent.
	parentPath, err := siaPath.Dir()
	if err != nil {
		return err
	}
	parent, err := fs.managedOpenDir(parentPath.String())
	if errors.Contains(err, ErrNotExist) {
		// If the parent doesn't exist yet we create it.
		err = fs.managedNewTurtleDexDir(parentPath, mode)
		if err == nil {
			parent, err = fs.managedOpenDir(parentPath.String())
		}
	}
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Compose(err, parent.Close())
	}()
	// Create the dir within the parent.
	return parent.managedNewTurtleDexDir(siaPath.Name(), fs.managedAbsPath(), mode)
}

// managedOpenFile opens a TurtleDexFile and adds it and all of its parents to the
// filesystem tree.
func (fs *FileSystem) managedOpenFile(relPath string) (_ *FileNode, err error) {
	// Open the folder that contains the file.
	dirPath, fileName := filepath.Split(relPath)
	var dir *DirNode
	if dirPath == string(filepath.Separator) || dirPath == "." || dirPath == "" {
		dir = &fs.DirNode // file is in the root dir
	} else {
		var err error
		dir, err = fs.managedOpenDir(filepath.Dir(relPath))
		if err != nil {
			return nil, errors.AddContext(err, "failed to open parent dir of file")
		}
		// Close the dir since we are not returning it. The open file keeps it
		// loaded in memory.
		defer func() {
			err = errors.Compose(err, dir.Close())
		}()
	}
	return dir.managedOpenFile(fileName)
}

// managedNewTurtleDexFile opens the parent folder of the new TurtleDexFile and calls
// managedNewTurtleDexFile on it.
func (fs *FileSystem) managedNewTurtleDexFile(relPath string, source string, ec modules.ErasureCoder, mk crypto.CipherKey, fileSize uint64, fileMode os.FileMode, disablePartialUpload bool) (err error) {
	// Open the folder that contains the file.
	dirPath, fileName := filepath.Split(relPath)
	var dir *DirNode
	if dirPath == string(filepath.Separator) || dirPath == "." || dirPath == "" {
		dir = &fs.DirNode // file is in the root dir
	} else {
		var err error
		dir, err = fs.managedOpenDir(filepath.Dir(relPath))
		if err != nil {
			return errors.AddContext(err, "failed to open parent dir of new file")
		}
		defer func() {
			err = errors.Compose(err, dir.Close())
		}()
	}
	return dir.managedNewTurtleDexFile(fileName, source, ec, mk, fileSize, fileMode, disablePartialUpload)
}

// managedOpenTurtleDexDir opens a TurtleDexDir and adds it and all of its parents to the
// filesystem tree.
func (fs *FileSystem) managedOpenTurtleDexDir(siaPath modules.TurtleDexPath) (*DirNode, error) {
	if siaPath.IsRoot() {
		// Make sure the metadata exists.
		_, err := os.Stat(filepath.Join(fs.absPath(), modules.TurtleDexDirExtension))
		if os.IsNotExist(err) {
			return nil, ErrNotExist
		}
		return fs.DirNode.managedCopy(), nil
	}
	dir, err := fs.DirNode.managedOpenDir(siaPath.String())
	if err != nil {
		return nil, err
	}
	return dir, nil
}
