package siatest

import (
	"github.com/turtledex/TurtleDexCore/modules"
)

type (
	// RemoteDir is a helper struct that represents a directory on the TurtleDex
	// network.
	RemoteDir struct {
		siapath modules.TurtleDexPath
	}
)

// TurtleDexPath returns the siapath of a remote directory.
func (rd *RemoteDir) TurtleDexPath() modules.TurtleDexPath {
	return rd.siapath
}
