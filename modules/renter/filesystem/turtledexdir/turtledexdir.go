package ttdxdir

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/turtledex/writeaheadlog"

	"github.com/turtledex/TurtleDexCore/modules"
)

type (
	// TurtleDexDir contains the metadata information about a renter directory
	TurtleDexDir struct {
		metadata Metadata

		// path is the path of the TurtleDexDir folder.
		path string

		// Utility fields
		deleted bool
		deps    modules.Dependencies
		mu      sync.Mutex
		wal     *writeaheadlog.WAL
	}

	// Metadata is the metadata that is saved to disk as a .ttdxdir file
	Metadata struct {
		// For each field in the metadata there is an aggregate value and a
		// ttdxdir specific value. If a field has the aggregate prefix it means
		// that the value takes into account all the siafiles and ttdxdirs in the
		// sub tree. The definition of aggregate and ttdxdir specific values is
		// otherwise the same.
		//
		// Health is the health of the most in need siafile that is not stuck
		//
		// LastHealthCheckTime is the oldest LastHealthCheckTime of any of the
		// siafiles in the ttdxdir and is the last time the health was calculated
		// by the health loop
		//
		// MinRedundancy is the minimum redundancy of any of the siafiles in the
		// ttdxdir
		//
		// ModTime is the last time any of the siafiles in the ttdxdir was
		// updated
		//
		// NumFiles is the total number of siafiles in a ttdxdir
		//
		// NumStuckChunks is the sum of all the Stuck Chunks of any of the
		// siafiles in the ttdxdir
		//
		// NumSubDirs is the number of sub-ttdxdirs in a ttdxdir
		//
		// Size is the total amount of data stored in the siafiles of the ttdxdir
		//
		// StuckHealth is the health of the most in need siafile in the ttdxdir,
		// stuck or not stuck

		// The following fields are aggregate values of the ttdxdir. These values are
		// the totals of the ttdxdir and any sub ttdxdirs, or are calculated based on
		// all the values in the subtree
		AggregateHealth              float64   `json:"aggregatehealth"`
		AggregateLastHealthCheckTime time.Time `json:"aggregatelasthealthchecktime"`
		AggregateMinRedundancy       float64   `json:"aggregateminredundancy"`
		AggregateModTime             time.Time `json:"aggregatemodtime"`
		AggregateNumFiles            uint64    `json:"aggregatenumfiles"`
		AggregateNumStuckChunks      uint64    `json:"aggregatenumstuckchunks"`
		AggregateNumSubDirs          uint64    `json:"aggregatenumsubdirs"`
		AggregateRemoteHealth        float64   `json:"aggregateremotehealth"`
		AggregateRepairSize          uint64    `json:"aggregaterepairsize"`
		AggregateSize                uint64    `json:"aggregatesize"`
		AggregateStuckHealth         float64   `json:"aggregatestuckhealth"`
		AggregateStuckSize           uint64    `json:"aggregatestucksize"`

		// Aggregate Skynet Specific Stats
		AggregateSkynetFiles uint64 `json:"aggregateskynetfiles"`
		AggregateSkynetSize  uint64 `json:"aggregateskynetsize"`

		// The following fields are information specific to the ttdxdir that is not
		// an aggregate of the entire sub directory tree
		Health              float64     `json:"health"`
		LastHealthCheckTime time.Time   `json:"lasthealthchecktime"`
		MinRedundancy       float64     `json:"minredundancy"`
		Mode                os.FileMode `json:"mode"`
		ModTime             time.Time   `json:"modtime"`
		NumFiles            uint64      `json:"numfiles"`
		NumStuckChunks      uint64      `json:"numstuckchunks"`
		NumSubDirs          uint64      `json:"numsubdirs"`
		RemoteHealth        float64     `json:"remotehealth"`
		RepairSize          uint64      `json:"repairsize"`
		Size                uint64      `json:"size"`
		StuckHealth         float64     `json:"stuckhealth"`
		StuckSize           uint64      `json:"stucksize"`

		// Skynet Specific Stats
		SkynetFiles uint64 `json:"skynetfiles"`
		SkynetSize  uint64 `json:"skynetsize"`

		// Version is the used version of the header file.
		Version string `json:"version"`
	}
)

// mdPath returns the path of the TurtleDexDir's metadata on disk.
func (sd *TurtleDexDir) mdPath() string {
	return filepath.Join(sd.path, modules.TurtleDexDirExtension)
}

// Deleted returns the deleted field of the siaDir
func (sd *TurtleDexDir) Deleted() bool {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.deleted
}

// Metadata returns the metadata of the TurtleDexDir
func (sd *TurtleDexDir) Metadata() Metadata {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.metadata
}

// Path returns the path of the TurtleDexDir on disk.
func (sd *TurtleDexDir) Path() string {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.path
}

// MDPath returns the path of the TurtleDexDir's metadata on disk.
func (sd *TurtleDexDir) MDPath() string {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.mdPath()
}
