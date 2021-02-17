package siafile

import (
	"path/filepath"

	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/errors"
	"github.com/turtledex/writeaheadlog"
)

// CombinedChunkIndex is a helper method which translates a chunk's index to the
// corresponding combined chunk index dependng on the number of combined chunks.
func CombinedChunkIndex(numChunks, chunkIndex uint64, numCombinedChunks int) int {
	if numCombinedChunks == 1 && chunkIndex == numChunks-1 {
		return 0
	}
	if numCombinedChunks == 2 && chunkIndex == numChunks-2 {
		return 0
	}
	if numCombinedChunks == 2 && chunkIndex == numChunks-1 {
		return 1
	}
	return -1
}

// Merge merges two PartialsTurtleDexfiles into one, returning a map which translates
// chunk indices in newFile to indices in sf.
func (sf *TurtleDexFile) Merge(newFile *TurtleDexFile) (map[uint64]uint64, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.merge(newFile)
}

// addCombinedChunk adds a new combined chunk to a combined TurtleDexfile. This can't
// be called on a regular TurtleDexFile.
func (sf *TurtleDexFile) addCombinedChunk() ([]writeaheadlog.Update, error) {
	if sf.deleted {
		return nil, errors.New("can't add combined chunk to deleted file")
	}
	if filepath.Ext(sf.siaFilePath) != modules.PartialsTurtleDexFileExtension {
		return nil, errors.New("can only call addCombinedChunk on combined TurtleDexFiles")
	}
	// Create updates to add a chunk and return index of that new chunk.
	updates, err := sf.growNumChunks(uint64(sf.numChunks) + 1)
	return updates, err
}

// merge merges two PartialsTurtleDexfiles into one, returning a map which translates
// chunk indices in newFile to indices in sf.
func (sf *TurtleDexFile) merge(newFile *TurtleDexFile) (map[uint64]uint64, error) {
	if sf.deleted {
		return nil, errors.New("can't merge into deleted file")
	}
	if filepath.Ext(sf.siaFilePath) != modules.PartialsTurtleDexFileExtension {
		return nil, errors.New("can only call merge on PartialsTurtleDexFile")
	}
	if filepath.Ext(newFile.TurtleDexFilePath()) != modules.PartialsTurtleDexFileExtension {
		return nil, errors.New("can only merge PartialsTurtleDexfiles into a PartialsTurtleDexFile")
	}
	newFile.mu.Lock()
	defer newFile.mu.Unlock()
	if newFile.deleted {
		return nil, errors.New("can't merge deleted file")
	}
	var newChunks []chunk
	indexMap := make(map[uint64]uint64)
	ncb := sf.numChunks
	err := newFile.iterateChunksReadonly(func(chunk chunk) error {
		newIndex := sf.numChunks
		indexMap[uint64(chunk.Index)] = uint64(newIndex)
		chunk.Index = newIndex
		newChunks = append(newChunks, chunk)
		return nil
	})
	if err != nil {
		sf.numChunks = ncb
		return nil, err
	}
	return indexMap, sf.saveFile(newChunks)
}
