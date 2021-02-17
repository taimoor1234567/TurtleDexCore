package siatest

import (
	"sync"

	"github.com/turtledex/TurtleDexCore/crypto"
	"github.com/turtledex/TurtleDexCore/modules"
)

type (
	// RemoteFile is a helper struct that represents a file uploaded to the TurtleDex
	// network.
	RemoteFile struct {
		checksum crypto.Hash
		siaPath  modules.TurtleDexPath
		root     bool
		mu       sync.Mutex
	}
)

// Checksum returns the checksum of a remote file.
func (rf *RemoteFile) Checksum() crypto.Hash {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.checksum
}

// Root returns whether the siapath needs to be treated as an absolute path.
func (rf *RemoteFile) Root() bool {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.root
}

// TurtleDexPath returns the siaPath of a remote file.
func (rf *RemoteFile) TurtleDexPath() modules.TurtleDexPath {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.siaPath
}
