package explorer

import (
	"github.com/turtledex/bolt"

	"github.com/turtledex/TurtleDexCore/build"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
)

// Block takes a block ID and finds the corresponding block, provided that the
// block is in the consensus set.
func (e *Explorer) Block(id types.BlockID) (types.Block, types.BlockHeight, bool) {
	var height types.BlockHeight
	err := e.db.View(dbGetAndDecode(bucketBlockIDs, id, &height))
	if err != nil {
		return types.Block{}, 0, false
	}
	block, exists := e.cs.BlockAtHeight(height)
	if !exists {
		return types.Block{}, 0, false
	}
	return block, height, true
}

// BlockFacts returns a set of statistics about the blockchain as they appeared
// at a given block height, and a bool indicating whether facts exist for the
// given height.
func (e *Explorer) BlockFacts(height types.BlockHeight) (modules.BlockFacts, bool) {
	var bf blockFacts
	err := e.db.View(e.dbGetBlockFacts(height, &bf))
	if err != nil {
		return modules.BlockFacts{}, false
	}

	return bf.BlockFacts, true
}

// LatestBlockFacts returns a set of statistics about the blockchain as they appeared
// at the latest block height in the explorer's consensus set.
func (e *Explorer) LatestBlockFacts() modules.BlockFacts {
	var bf blockFacts
	err := e.db.View(func(tx *bolt.Tx) error {
		var height types.BlockHeight
		err := dbGetInternal(internalBlockHeight, &height)(tx)
		if err != nil {
			return err
		}
		return e.dbGetBlockFacts(height, &bf)(tx)
	})
	if err != nil {
		build.Critical(err)
	}
	return bf.BlockFacts
}

// Transaction takes a transaction ID and finds the block containing the
// transaction. Because of the miner payouts, the transaction ID might be a
// block ID. To find the transaction, iterate through the block.
func (e *Explorer) Transaction(id types.TransactionID) (types.Block, types.BlockHeight, bool) {
	var height types.BlockHeight
	err := e.db.View(dbGetAndDecode(bucketTransactionIDs, id, &height))
	if err != nil {
		return types.Block{}, 0, false
	}
	block, exists := e.cs.BlockAtHeight(height)
	if !exists {
		return types.Block{}, 0, false
	}
	return block, height, true
}

// UnlockHash returns the IDs of all the transactions that contain the unlock
// hash. An empty set indicates that the unlock hash does not appear in the
// blockchain.
func (e *Explorer) UnlockHash(uh types.UnlockHash) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketUnlockHashes, uh, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}

// TurtleDexcoinOutput returns the ttdc output associated with the specified ID.
func (e *Explorer) TurtleDexcoinOutput(id types.TurtleDexcoinOutputID) (types.TurtleDexcoinOutput, bool) {
	var sco types.TurtleDexcoinOutput
	err := e.db.View(dbGetAndDecode(bucketTurtleDexcoinOutputs, id, &sco))
	if err != nil {
		return types.TurtleDexcoinOutput{}, false
	}
	return sco, true
}

// TurtleDexcoinOutputID returns all of the transactions that contain the specified
// ttdc output ID. An empty set indicates that the ttdc output ID does
// not appear in the blockchain.
func (e *Explorer) TurtleDexcoinOutputID(id types.TurtleDexcoinOutputID) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketTurtleDexcoinOutputIDs, id, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}

// FileContractHistory returns the history associated with the specified file
// contract ID, which includes the file contract itself and all of the
// revisions that have been submitted to the blockchain. The first bool
// indicates whether the file contract exists, and the second bool indicates
// whether a storage proof was successfully submitted for the file contract.
func (e *Explorer) FileContractHistory(id types.FileContractID) (fc types.FileContract, fcrs []types.FileContractRevision, fcE bool, spE bool) {
	var history fileContractHistory
	err := e.db.View(dbGetAndDecode(bucketFileContractHistories, id, &history))
	fc = history.Contract
	fcrs = history.Revisions
	fcE = err == nil
	spE = history.StorageProof.ParentID == id
	return
}

// FileContractID returns all transactions that contain the specified
// file contract ID. An empty set indicates that the file contract ID does not
// appear in the blockchain.
func (e *Explorer) FileContractID(id types.FileContractID) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketFileContractIDs, id, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}

// FileContractPayouts returns all of the spendable ttdc outputs which are the
// result of a FileContract. An empty set indicates that the file contract is
// still open
func (e *Explorer) FileContractPayouts(id types.FileContractID) ([]types.TurtleDexcoinOutput, error) {
	var history fileContractHistory
	err := e.db.View(dbGetAndDecode(bucketFileContractHistories, id, &history))
	if err != nil {
		return []types.TurtleDexcoinOutput{}, err
	}

	fc := history.Contract
	var outputs []types.TurtleDexcoinOutput

	for i := range fc.ValidProofOutputs {
		scoid := id.StorageProofOutputID(types.ProofValid, uint64(i))

		sco, found := e.TurtleDexcoinOutput(scoid)
		if found {
			outputs = append(outputs, sco)
		}
	}
	for i := range fc.MissedProofOutputs {
		scoid := id.StorageProofOutputID(types.ProofMissed, uint64(i))

		sco, found := e.TurtleDexcoinOutput(scoid)
		if found {
			outputs = append(outputs, sco)
		}
	}

	return outputs, nil
}

// TurtleDexfundOutput returns the siafund output associated with the specified ID.
func (e *Explorer) TurtleDexfundOutput(id types.TurtleDexfundOutputID) (types.TurtleDexfundOutput, bool) {
	var sco types.TurtleDexfundOutput
	err := e.db.View(dbGetAndDecode(bucketTurtleDexfundOutputs, id, &sco))
	if err != nil {
		return types.TurtleDexfundOutput{}, false
	}
	return sco, true
}

// TurtleDexfundOutputID returns all of the transactions that contain the specified
// siafund output ID. An empty set indicates that the siafund output ID does
// not appear in the blockchain.
func (e *Explorer) TurtleDexfundOutputID(id types.TurtleDexfundOutputID) []types.TransactionID {
	var ids []types.TransactionID
	err := e.db.View(dbGetTransactionIDSet(bucketTurtleDexfundOutputIDs, id, &ids))
	if err != nil {
		ids = nil
	}
	return ids
}
