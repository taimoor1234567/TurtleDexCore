package modules

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/turtledex/errors"
	"github.com/turtledex/fastrand"
)

// siapath.go contains the types and methods for creating and manipulating
// siapaths. Any methods such as filepath.Join should be implemented here for
// the TurtleDexPath type to ensure consistent handling across OS.

var (
	// ErrEmptyPath is an error when a path is empty
	ErrEmptyPath = errors.New("path must be a nonempty string")
	// ErrInvalidTurtleDexPath is the error for an invalid TurtleDexPath
	ErrInvalidTurtleDexPath = errors.New("invalid TurtleDexPath")
	// ErrInvalidPathString is the error for an invalid path
	ErrInvalidPathString = errors.New("invalid path string")

	// TurtleDexDirExtension is the extension for ttdxdir metadata files on disk
	TurtleDexDirExtension = ".ttdxdir"

	// TurtleDexFileExtension is the extension for siafiles on disk
	TurtleDexFileExtension = ".sia"

	// PartialsTurtleDexFileExtension is the extension for siafiles which contain
	// combined chunks.
	PartialsTurtleDexFileExtension = ".csia"

	// CombinedChunkExtension is the extension for a combined chunk on disk.
	CombinedChunkExtension = ".cc"
	// UnfinishedChunkExtension is the extension for an unfinished combined chunk
	// and is appended to the file in addition to CombinedChunkExtension.
	UnfinishedChunkExtension = ".unfinished"
	// ChunkMetadataExtension is the extension of a metadata file for a combined
	// chunk.
	ChunkMetadataExtension = ".ccmd"
)

var (
	// BackupFolder is the TurtleDex folder where all of the renter's snapshot
	// siafiles are stored by default.
	BackupFolder = NewGlobalTurtleDexPath("/snapshots")

	// HomeFolder is the TurtleDex folder that is used to store all of the user
	// accessible data.
	HomeFolder = NewGlobalTurtleDexPath("/home")

	// SkynetFolder is the TurtleDex folder where all of the skyfiles are stored by
	// default.
	SkynetFolder = NewGlobalTurtleDexPath("/var/skynet")

	// UserFolder is the TurtleDex folder that is used to store the renter's siafiles.
	UserFolder = NewGlobalTurtleDexPath("/home/user")

	// VarFolder is the TurtleDex folder that contains the skynet folder.
	VarFolder = NewGlobalTurtleDexPath("/var")
)

type (
	// TurtleDexPath is the struct used to uniquely identify siafiles and ttdxdirs across
	// TurtleDex
	TurtleDexPath struct {
		Path string `json:"path"`
	}
)

// NewTurtleDexPath returns a new TurtleDexPath with the path set
func NewTurtleDexPath(s string) (TurtleDexPath, error) {
	return newTurtleDexPath(s)
}

// NewGlobalTurtleDexPath can be used to create a global var which is a TurtleDexPath. If
// there is an error creating the TurtleDexPath, the function will panic, making this
// function unsuitable for typical use.
func NewGlobalTurtleDexPath(s string) TurtleDexPath {
	sp, err := NewTurtleDexPath(s)
	if err != nil {
		panic("error creating global siapath: " + err.Error())
	}
	return sp
}

// RandomTurtleDexPath returns a random TurtleDexPath created from 20 bytes of base32
// encoded entropy.
func RandomTurtleDexPath() (sp TurtleDexPath) {
	sp.Path = base32.StdEncoding.EncodeToString(fastrand.Bytes(20))
	sp.Path = sp.Path[:20]
	return
}

// RootTurtleDexPath returns a TurtleDexPath for the root ttdxdir which has a blank path
func RootTurtleDexPath() TurtleDexPath {
	return TurtleDexPath{}
}

// CombinedTurtleDexFilePath returns the TurtleDexPath to a hidden siafile which is used to
// store chunks that contain pieces of multiple siafiles.
func CombinedTurtleDexFilePath(ec ErasureCoder) TurtleDexPath {
	return TurtleDexPath{Path: fmt.Sprintf(".%v", ec.Identifier())}
}

// clean cleans up the string by converting an OS separators to forward slashes
// and trims leading and trailing slashes
func clean(s string) string {
	s = filepath.ToSlash(s)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	return s
}

// newTurtleDexPath returns a new TurtleDexPath with the path set
func newTurtleDexPath(s string) (TurtleDexPath, error) {
	sp := TurtleDexPath{
		Path: clean(s),
	}
	return sp, sp.Validate(false)
}

// AddSuffix adds a numeric suffix to the end of the TurtleDexPath.
func (sp TurtleDexPath) AddSuffix(suffix uint) TurtleDexPath {
	return TurtleDexPath{
		Path: sp.Path + fmt.Sprintf("_%v", suffix),
	}
}

// Dir returns the directory of the TurtleDexPath
func (sp TurtleDexPath) Dir() (TurtleDexPath, error) {
	pathElements := strings.Split(sp.Path, "/")
	// If there is only one path element, then the TurtleDexpath was just a filename
	// and did not have a directory, return the root TurtleDexpath
	if len(pathElements) <= 1 {
		return RootTurtleDexPath(), nil
	}
	dir := strings.Join(pathElements[:len(pathElements)-1], "/")
	// If dir is empty or a dot, return the root TurtleDexpath
	if dir == "" || dir == "." {
		return RootTurtleDexPath(), nil
	}
	return newTurtleDexPath(dir)
}

// Equals compares two TurtleDexPath types for equality
func (sp TurtleDexPath) Equals(siaPath TurtleDexPath) bool {
	return sp.Path == siaPath.Path
}

// IsEmpty returns true if the siapath is equal to the nil value
func (sp TurtleDexPath) IsEmpty() bool {
	return sp.Equals(TurtleDexPath{})
}

// IsRoot indicates whether or not the TurtleDexPath path is a root directory siapath
func (sp TurtleDexPath) IsRoot() bool {
	return sp.Path == ""
}

// Join joins the string to the end of the TurtleDexPath with a "/" and returns the
// new TurtleDexPath.
func (sp TurtleDexPath) Join(s string) (TurtleDexPath, error) {
	cleanStr := clean(s)
	if s == "" || cleanStr == "" {
		return TurtleDexPath{}, errors.New("cannot join an empty string to a siapath")
	}
	return newTurtleDexPath(sp.Path + "/" + cleanStr)
}

// LoadString sets the path of the TurtleDexPath to the provided string
func (sp *TurtleDexPath) LoadString(s string) error {
	sp.Path = clean(s)
	return sp.Validate(false)
}

// LoadSysPath loads a TurtleDexPath from a given system path by trimming the dir at
// the front of the path, the extension at the back and returning the remaining
// path as a TurtleDexPath.
func (sp *TurtleDexPath) LoadSysPath(dir, path string) error {
	if !strings.HasPrefix(path, dir) {
		return fmt.Errorf("%v is not a prefix of %v", dir, path)
	}
	path = strings.TrimSuffix(strings.TrimPrefix(path, dir), TurtleDexFileExtension)
	return sp.LoadString(path)
}

// MarshalJSON marshals a TurtleDexPath as a string.
func (sp TurtleDexPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(sp.String())
}

// Name returns the name of the file.
func (sp TurtleDexPath) Name() string {
	pathElements := strings.Split(sp.Path, "/")
	name := pathElements[len(pathElements)-1]
	// If name is a dot, return the root TurtleDexpath name
	if name == "." {
		name = ""
	}
	return name
}

// Rebase changes the base of a siapath from oldBase to newBase and returns a new TurtleDexPath.
// e.g. rebasing 'a/b/myfile' from oldBase 'a/b/' to 'a/' would result in 'a/myfile'
func (sp TurtleDexPath) Rebase(oldBase, newBase TurtleDexPath) (TurtleDexPath, error) {
	if !strings.HasPrefix(sp.Path, oldBase.Path) {
		return TurtleDexPath{}, fmt.Errorf("'%v' isn't the base of '%v'", oldBase.Path, sp.Path)
	}
	relPath := strings.TrimPrefix(sp.Path, oldBase.Path)
	if relPath == "" {
		return newBase, nil
	}
	return newBase.Join(relPath)
}

// UnmarshalJSON unmarshals a siapath into a TurtleDexPath object.
func (sp *TurtleDexPath) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &sp.Path); err != nil {
		return err
	}
	sp.Path = clean(sp.Path)
	return sp.Validate(true)
}

// TurtleDexDirSysPath returns the system path needed to read a directory on disk, the
// input dir is the root ttdxdir directory on disk
func (sp TurtleDexPath) TurtleDexDirSysPath(dir string) string {
	return filepath.Join(dir, filepath.FromSlash(sp.Path), "")
}

// TurtleDexDirMetadataSysPath returns the system path needed to read the TurtleDexDir
// metadata file from disk, the input dir is the root ttdxdir directory on disk
func (sp TurtleDexPath) TurtleDexDirMetadataSysPath(dir string) string {
	return filepath.Join(dir, filepath.FromSlash(sp.Path), TurtleDexDirExtension)
}

// TurtleDexFileSysPath returns the system path needed to read the TurtleDexFile from disk,
// the input dir is the root siafile directory on disk
func (sp TurtleDexPath) TurtleDexFileSysPath(dir string) string {
	return filepath.Join(dir, filepath.FromSlash(sp.Path)+TurtleDexFileExtension)
}

// TurtleDexPartialsFileSysPath returns the system path needed to read the
// PartialsTurtleDexFile from disk, the input dir is the root siafile directory on
// disk
func (sp TurtleDexPath) TurtleDexPartialsFileSysPath(dir string) string {
	return filepath.Join(dir, filepath.FromSlash(sp.Path)+PartialsTurtleDexFileExtension)
}

// String returns the TurtleDexPath's path
func (sp TurtleDexPath) String() string {
	return sp.Path
}

// FromSysPath creates a TurtleDexPath from a siaFilePath and corresponding root files
// dir.
func (sp *TurtleDexPath) FromSysPath(siaFilePath, dir string) (err error) {
	if !strings.HasPrefix(siaFilePath, dir) {
		return fmt.Errorf("TurtleDexFilePath %v is not within dir %v", siaFilePath, dir)
	}
	relPath := strings.TrimPrefix(siaFilePath, dir)
	relPath = strings.TrimSuffix(relPath, TurtleDexFileExtension)
	relPath = strings.TrimSuffix(relPath, PartialsTurtleDexFileExtension)
	*sp, err = newTurtleDexPath(relPath)
	return
}

// Validate checks that a TurtleDexpath is a legal filename.
func (sp TurtleDexPath) Validate(isRoot bool) error {
	if err := validatePath(sp.Path, isRoot); err != nil {
		return errors.Extend(err, ErrInvalidTurtleDexPath)
	}
	return nil
}

// ValidatePathString validates a path given a string.
func ValidatePathString(path string, isRoot bool) error {
	if err := validatePath(path, isRoot); err != nil {
		return errors.Extend(err, ErrInvalidPathString)
	}
	return nil
}

// validatePath validates a path. ../ and ./ are disallowed to prevent directory
// traversal, and paths must not begin with / or be empty.
func validatePath(path string, isRoot bool) error {
	if path == "" && !isRoot {
		return ErrEmptyPath
	}
	if path == ".." {
		return errors.New("path cannot be '..'")
	}
	if path == "." {
		return errors.New("path cannot be '.'")
	}
	// check prefix
	if strings.HasPrefix(path, "/") {
		return errors.New("path cannot begin with /")
	}
	if strings.HasPrefix(path, "../") {
		return errors.New("path cannot begin with ../")
	}
	if strings.HasPrefix(path, "./") {
		return errors.New("path connot begin with ./")
	}
	var prevElem string
	for _, pathElem := range strings.Split(path, "/") {
		if pathElem == "." || pathElem == ".." {
			return errors.New("path cannot contain . or .. elements")
		}
		if prevElem != "" && pathElem == "" {
			return ErrEmptyPath
		}
		if prevElem == "/" || pathElem == "/" {
			return errors.New("path cannot contain //")
		}
		prevElem = pathElem
	}

	// Final check for a valid utf8
	if !utf8.ValidString(path) {
		return errors.New("path is not a valid utf8 path")
	}

	return nil
}
