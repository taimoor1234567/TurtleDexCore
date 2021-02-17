package explorer

import (
	"fmt"

	"github.com/turtledex/bolt"

	"github.com/turtledex/TurtleDexCore/build"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
	"github.com/turtledex/encoding"
)

// ProcessConsensusChange follows the most recent changes to the consensus set,
// including parsing new blocks and updating the utxo sets.
func (e *Explorer) ProcessConsensusChange(cc modules.ConsensusChange) {
	if len(cc.AppliedBlocks) == 0 {
		build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
	}

	err := e.db.Update(func(tx *bolt.Tx) (err error) {
		// use exception-style error handling to enable more concise update code
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()

		// get starting block height
		var blockheight types.BlockHeight
		err = dbGetInternal(internalBlockHeight, &blockheight)(tx)
		if err != nil {
			return err
		}

		// Update cumulative stats for reverted blocks.
		for _, block := range cc.RevertedBlocks {
			bid := block.ID()
			tbid := types.TransactionID(bid)

			blockheight--
			dbRemoveBlockID(tx, bid)
			dbRemoveTransactionID(tx, tbid) // Miner payouts are a transaction

			target, exists := e.cs.ChildTarget(block.ParentID)
			if !exists {
				target = types.RootTarget
			}
			dbRemoveBlockTarget(tx, bid, target)

			// Remove miner payouts
			for j, payout := range block.MinerPayouts {
				scoid := block.MinerPayoutID(uint64(j))
				dbRemoveTurtleDexcoinOutputID(tx, scoid, tbid)
				dbRemoveUnlockHash(tx, payout.UnlockHash, tbid)
			}

			// Remove transactions
			for _, txn := range block.Transactions {
				txid := txn.ID()
				dbRemoveTransactionID(tx, txid)

				for _, sci := range txn.TurtleDexcoinInputs {
					dbRemoveTurtleDexcoinOutputID(tx, sci.ParentID, txid)
					dbRemoveUnlockHash(tx, sci.UnlockConditions.UnlockHash(), txid)
				}
				for k, sco := range txn.TurtleDexcoinOutputs {
					scoid := txn.TurtleDexcoinOutputID(uint64(k))
					dbRemoveTurtleDexcoinOutputID(tx, scoid, txid)
					dbRemoveUnlockHash(tx, sco.UnlockHash, txid)
					dbRemoveTurtleDexcoinOutput(tx, scoid)
				}
				for k, fc := range txn.FileContracts {
					fcid := txn.FileContractID(uint64(k))
					dbRemoveFileContractID(tx, fcid, txid)
					dbRemoveUnlockHash(tx, fc.UnlockHash, txid)
					for l, sco := range fc.ValidProofOutputs {
						scoid := fcid.StorageProofOutputID(types.ProofValid, uint64(l))
						dbRemoveTurtleDexcoinOutputID(tx, scoid, txid)
						dbRemoveUnlockHash(tx, sco.UnlockHash, txid)
					}
					for l, sco := range fc.MissedProofOutputs {
						scoid := fcid.StorageProofOutputID(types.ProofMissed, uint64(l))
						dbRemoveTurtleDexcoinOutputID(tx, scoid, txid)
						dbRemoveUnlockHash(tx, sco.UnlockHash, txid)
					}
					dbRemoveFileContract(tx, fcid)
				}
				for _, fcr := range txn.FileContractRevisions {
					dbRemoveFileContractID(tx, fcr.ParentID, txid)
					dbRemoveUnlockHash(tx, fcr.UnlockConditions.UnlockHash(), txid)
					dbRemoveUnlockHash(tx, fcr.NewUnlockHash, txid)
					for l, sco := range fcr.NewValidProofOutputs {
						scoid := fcr.ParentID.StorageProofOutputID(types.ProofValid, uint64(l))
						dbRemoveTurtleDexcoinOutputID(tx, scoid, txid)
						dbRemoveUnlockHash(tx, sco.UnlockHash, txid)
					}
					for l, sco := range fcr.NewMissedProofOutputs {
						scoid := fcr.ParentID.StorageProofOutputID(types.ProofMissed, uint64(l))
						dbRemoveTurtleDexcoinOutputID(tx, scoid, txid)
						dbRemoveUnlockHash(tx, sco.UnlockHash, txid)
					}
					// Remove the file contract revision from the revision chain.
					dbRemoveFileContractRevision(tx, fcr.ParentID)
				}
				for _, sp := range txn.StorageProofs {
					dbRemoveStorageProof(tx, sp.ParentID)
				}
				for _, sfi := range txn.TurtleDexfundInputs {
					dbRemoveTurtleDexfundOutputID(tx, sfi.ParentID, txid)
					dbRemoveUnlockHash(tx, sfi.UnlockConditions.UnlockHash(), txid)
					dbRemoveUnlockHash(tx, sfi.ClaimUnlockHash, txid)
				}
				for k, sfo := range txn.TurtleDexfundOutputs {
					sfoid := txn.TurtleDexfundOutputID(uint64(k))
					dbRemoveTurtleDexfundOutputID(tx, sfoid, txid)
					dbRemoveUnlockHash(tx, sfo.UnlockHash, txid)
				}
			}

			// remove the associated block facts
			dbRemoveBlockFacts(tx, bid)
		}

		// Update cumulative stats for applied blocks.
		for _, block := range cc.AppliedBlocks {
			bid := block.ID()
			tbid := types.TransactionID(bid)

			// special handling for genesis block
			if bid == types.GenesisID {
				dbAddGenesisBlock(tx)
				continue
			}

			blockheight++
			dbAddBlockID(tx, bid, blockheight)
			dbAddTransactionID(tx, tbid, blockheight) // Miner payouts are a transaction

			target, exists := e.cs.ChildTarget(block.ParentID)
			if !exists {
				target = types.RootTarget
			}
			dbAddBlockTarget(tx, bid, target)

			// Catalog the new miner payouts.
			for j, payout := range block.MinerPayouts {
				scoid := block.MinerPayoutID(uint64(j))
				dbAddTurtleDexcoinOutputID(tx, scoid, tbid)
				dbAddUnlockHash(tx, payout.UnlockHash, tbid)
			}

			// Update cumulative stats for applied transactions.
			for _, txn := range block.Transactions {
				// Add the transaction to the list of active transactions.
				txid := txn.ID()
				dbAddTransactionID(tx, txid, blockheight)

				for _, sci := range txn.TurtleDexcoinInputs {
					dbAddTurtleDexcoinOutputID(tx, sci.ParentID, txid)
					dbAddUnlockHash(tx, sci.UnlockConditions.UnlockHash(), txid)
				}
				for j, sco := range txn.TurtleDexcoinOutputs {
					scoid := txn.TurtleDexcoinOutputID(uint64(j))
					dbAddTurtleDexcoinOutputID(tx, scoid, txid)
					dbAddUnlockHash(tx, sco.UnlockHash, txid)
				}
				for k, fc := range txn.FileContracts {
					fcid := txn.FileContractID(uint64(k))
					dbAddFileContractID(tx, fcid, txid)
					dbAddUnlockHash(tx, fc.UnlockHash, txid)
					dbAddFileContract(tx, fcid, fc)
					for l, sco := range fc.ValidProofOutputs {
						scoid := fcid.StorageProofOutputID(types.ProofValid, uint64(l))
						dbAddTurtleDexcoinOutputID(tx, scoid, txid)
						dbAddUnlockHash(tx, sco.UnlockHash, txid)
					}
					for l, sco := range fc.MissedProofOutputs {
						scoid := fcid.StorageProofOutputID(types.ProofMissed, uint64(l))
						dbAddTurtleDexcoinOutputID(tx, scoid, txid)
						dbAddUnlockHash(tx, sco.UnlockHash, txid)
					}
				}
				for _, fcr := range txn.FileContractRevisions {
					dbAddFileContractID(tx, fcr.ParentID, txid)
					dbAddUnlockHash(tx, fcr.UnlockConditions.UnlockHash(), txid)
					dbAddUnlockHash(tx, fcr.NewUnlockHash, txid)
					for l, sco := range fcr.NewValidProofOutputs {
						scoid := fcr.ParentID.StorageProofOutputID(types.ProofValid, uint64(l))
						dbAddTurtleDexcoinOutputID(tx, scoid, txid)
						dbAddUnlockHash(tx, sco.UnlockHash, txid)
					}
					for l, sco := range fcr.NewMissedProofOutputs {
						scoid := fcr.ParentID.StorageProofOutputID(types.ProofMissed, uint64(l))
						dbAddTurtleDexcoinOutputID(tx, scoid, txid)
						dbAddUnlockHash(tx, sco.UnlockHash, txid)
					}
					dbAddFileContractRevision(tx, fcr.ParentID, fcr)
				}
				for _, sp := range txn.StorageProofs {
					dbAddFileContractID(tx, sp.ParentID, txid)
					dbAddStorageProof(tx, sp.ParentID, sp)
				}
				for _, sfi := range txn.TurtleDexfundInputs {
					dbAddTurtleDexfundOutputID(tx, sfi.ParentID, txid)
					dbAddUnlockHash(tx, sfi.UnlockConditions.UnlockHash(), txid)
					dbAddUnlockHash(tx, sfi.ClaimUnlockHash, txid)
				}
				for k, sfo := range txn.TurtleDexfundOutputs {
					sfoid := txn.TurtleDexfundOutputID(uint64(k))
					dbAddTurtleDexfundOutputID(tx, sfoid, txid)
					dbAddUnlockHash(tx, sfo.UnlockHash, txid)
				}
			}

			// calculate and add new block facts, if possible
			if tx.Bucket(bucketBlockFacts).Get(encoding.Marshal(block.ParentID)) != nil {
				facts := dbCalculateBlockFacts(tx, e.cs, block)
				dbAddBlockFacts(tx, facts)
			}
		}

		// Update stats according to TurtleDexcoinOutputDiffs
		for _, scod := range cc.TurtleDexcoinOutputDiffs {
			if scod.Direction == modules.DiffApply {
				dbAddTurtleDexcoinOutput(tx, scod.ID, scod.TurtleDexcoinOutput)
			}
		}

		// Update stats according to TurtleDexfundOutputDiffs
		for _, sfod := range cc.TurtleDexfundOutputDiffs {
			if sfod.Direction == modules.DiffApply {
				dbAddTurtleDexfundOutput(tx, sfod.ID, sfod.TurtleDexfundOutput)
			}
		}

		// Compute the changes in the active set. Note, because this is calculated
		// at the end instead of in a loop, the historic facts may contain
		// inaccuracies about the active set. This should not be a problem except
		// for large reorgs.
		// TODO: improve this
		currentBlock, exists := e.cs.BlockAtHeight(blockheight)
		if !exists {
			build.Critical("consensus is missing block", blockheight)
		}
		currentID := currentBlock.ID()
		var facts blockFacts
		err = dbGetAndDecode(bucketBlockFacts, currentID, &facts)(tx)
		if err == nil {
			for _, diff := range cc.FileContractDiffs {
				if diff.Direction == modules.DiffApply {
					facts.ActiveContractCount++
					facts.ActiveContractCost = facts.ActiveContractCost.Add(diff.FileContract.Payout)
					facts.ActiveContractSize = facts.ActiveContractSize.Add(types.NewCurrency64(diff.FileContract.FileSize))
				} else {
					facts.ActiveContractCount--
					facts.ActiveContractCost = facts.ActiveContractCost.Sub(diff.FileContract.Payout)
					facts.ActiveContractSize = facts.ActiveContractSize.Sub(types.NewCurrency64(diff.FileContract.FileSize))
				}
			}
			err = tx.Bucket(bucketBlockFacts).Put(encoding.Marshal(currentID), encoding.Marshal(facts))
			if err != nil {
				return err
			}
		}

		// set final blockheight
		err = dbSetInternal(internalBlockHeight, blockheight)(tx)
		if err != nil {
			return err
		}

		// set change ID
		err = dbSetInternal(internalRecentChange, cc.ID)(tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		build.Critical("explorer update failed:", err)
	}
}

// helper functions
func assertNil(err error) {
	if err != nil {
		panic(err)
	}
}
func mustPut(bucket *bolt.Bucket, key, val interface{}) {
	assertNil(bucket.Put(encoding.Marshal(key), encoding.Marshal(val)))
}
func mustPutSet(bucket *bolt.Bucket, key interface{}) {
	assertNil(bucket.Put(encoding.Marshal(key), nil))
}
func mustDelete(bucket *bolt.Bucket, key interface{}) {
	assertNil(bucket.Delete(encoding.Marshal(key)))
}
func bucketIsEmpty(bucket *bolt.Bucket) bool {
	k, _ := bucket.Cursor().First()
	return k == nil
}

// These functions panic on error. The panic will be caught by
// ProcessConsensusChange.

// Add/Remove block ID
func dbAddBlockID(tx *bolt.Tx, id types.BlockID, height types.BlockHeight) {
	mustPut(tx.Bucket(bucketBlockIDs), id, height)
}
func dbRemoveBlockID(tx *bolt.Tx, id types.BlockID) {
	mustDelete(tx.Bucket(bucketBlockIDs), id)
}

// Add/Remove block facts
func dbAddBlockFacts(tx *bolt.Tx, facts blockFacts) {
	mustPut(tx.Bucket(bucketBlockFacts), facts.BlockID, facts)
}
func dbRemoveBlockFacts(tx *bolt.Tx, id types.BlockID) {
	mustDelete(tx.Bucket(bucketBlockFacts), id)
}

// Add/Remove block target
func dbAddBlockTarget(tx *bolt.Tx, id types.BlockID, target types.Target) {
	mustPut(tx.Bucket(bucketBlockTargets), id, target)
}
func dbRemoveBlockTarget(tx *bolt.Tx, id types.BlockID, target types.Target) {
	mustDelete(tx.Bucket(bucketBlockTargets), id)
}

// Add/Remove file contract
func dbAddFileContract(tx *bolt.Tx, id types.FileContractID, fc types.FileContract) {
	history := fileContractHistory{Contract: fc}
	mustPut(tx.Bucket(bucketFileContractHistories), id, history)
}
func dbRemoveFileContract(tx *bolt.Tx, id types.FileContractID) {
	mustDelete(tx.Bucket(bucketFileContractHistories), id)
}

// Add/Remove txid from file contract ID bucket
func dbAddFileContractID(tx *bolt.Tx, id types.FileContractID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketFileContractIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveFileContractID(tx *bolt.Tx, id types.FileContractID, txid types.TransactionID) {
	bucket := tx.Bucket(bucketFileContractIDs).Bucket(encoding.Marshal(id))
	mustDelete(bucket, txid)
	if bucketIsEmpty(bucket) {
		tx.Bucket(bucketFileContractIDs).DeleteBucket(encoding.Marshal(id))
	}
}

func dbAddFileContractRevision(tx *bolt.Tx, fcid types.FileContractID, fcr types.FileContractRevision) {
	var history fileContractHistory
	assertNil(dbGetAndDecode(bucketFileContractHistories, fcid, &history)(tx))
	history.Revisions = append(history.Revisions, fcr)
	mustPut(tx.Bucket(bucketFileContractHistories), fcid, history)
}
func dbRemoveFileContractRevision(tx *bolt.Tx, fcid types.FileContractID) {
	var history fileContractHistory
	assertNil(dbGetAndDecode(bucketFileContractHistories, fcid, &history)(tx))
	// TODO: could be more rigorous
	history.Revisions = history.Revisions[:len(history.Revisions)-1]
	mustPut(tx.Bucket(bucketFileContractHistories), fcid, history)
}

// Add/Remove ttdc output
func dbAddTurtleDexcoinOutput(tx *bolt.Tx, id types.TurtleDexcoinOutputID, output types.TurtleDexcoinOutput) {
	mustPut(tx.Bucket(bucketTurtleDexcoinOutputs), id, output)
}
func dbRemoveTurtleDexcoinOutput(tx *bolt.Tx, id types.TurtleDexcoinOutputID) {
	mustDelete(tx.Bucket(bucketTurtleDexcoinOutputs), id)
}

// Add/Remove txid from ttdc output ID bucket
func dbAddTurtleDexcoinOutputID(tx *bolt.Tx, id types.TurtleDexcoinOutputID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketTurtleDexcoinOutputIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveTurtleDexcoinOutputID(tx *bolt.Tx, id types.TurtleDexcoinOutputID, txid types.TransactionID) {
	bucket := tx.Bucket(bucketTurtleDexcoinOutputIDs).Bucket(encoding.Marshal(id))
	mustDelete(bucket, txid)
	if bucketIsEmpty(bucket) {
		tx.Bucket(bucketTurtleDexcoinOutputIDs).DeleteBucket(encoding.Marshal(id))
	}
}

// Add/Remove siafund output
func dbAddTurtleDexfundOutput(tx *bolt.Tx, id types.TurtleDexfundOutputID, output types.TurtleDexfundOutput) {
	mustPut(tx.Bucket(bucketTurtleDexfundOutputs), id, output)
}

// Add/Remove txid from siafund output ID bucket
func dbAddTurtleDexfundOutputID(tx *bolt.Tx, id types.TurtleDexfundOutputID, txid types.TransactionID) {
	b, err := tx.Bucket(bucketTurtleDexfundOutputIDs).CreateBucketIfNotExists(encoding.Marshal(id))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveTurtleDexfundOutputID(tx *bolt.Tx, id types.TurtleDexfundOutputID, txid types.TransactionID) {
	bucket := tx.Bucket(bucketTurtleDexfundOutputIDs).Bucket(encoding.Marshal(id))
	mustDelete(bucket, txid)
	if bucketIsEmpty(bucket) {
		tx.Bucket(bucketTurtleDexfundOutputIDs).DeleteBucket(encoding.Marshal(id))
	}
}

// Add/Remove storage proof
func dbAddStorageProof(tx *bolt.Tx, fcid types.FileContractID, sp types.StorageProof) {
	var history fileContractHistory
	assertNil(dbGetAndDecode(bucketFileContractHistories, fcid, &history)(tx))
	history.StorageProof = sp
	mustPut(tx.Bucket(bucketFileContractHistories), fcid, history)
}
func dbRemoveStorageProof(tx *bolt.Tx, fcid types.FileContractID) {
	dbAddStorageProof(tx, fcid, types.StorageProof{})
}

// Add/Remove transaction ID
func dbAddTransactionID(tx *bolt.Tx, id types.TransactionID, height types.BlockHeight) {
	mustPut(tx.Bucket(bucketTransactionIDs), id, height)
}
func dbRemoveTransactionID(tx *bolt.Tx, id types.TransactionID) {
	mustDelete(tx.Bucket(bucketTransactionIDs), id)
}

// Add/Remove txid from unlock hash bucket
func dbAddUnlockHash(tx *bolt.Tx, uh types.UnlockHash, txid types.TransactionID) {
	b, err := tx.Bucket(bucketUnlockHashes).CreateBucketIfNotExists(encoding.Marshal(uh))
	assertNil(err)
	mustPutSet(b, txid)
}
func dbRemoveUnlockHash(tx *bolt.Tx, uh types.UnlockHash, txid types.TransactionID) {
	bucket := tx.Bucket(bucketUnlockHashes).Bucket(encoding.Marshal(uh))
	mustDelete(bucket, txid)
	if bucketIsEmpty(bucket) {
		tx.Bucket(bucketUnlockHashes).DeleteBucket(encoding.Marshal(uh))
	}
}

func dbCalculateBlockFacts(tx *bolt.Tx, cs modules.ConsensusSet, block types.Block) blockFacts {
	// get the parent block facts
	var bf blockFacts
	err := dbGetAndDecode(bucketBlockFacts, block.ParentID, &bf)(tx)
	assertNil(err)

	// get target
	target, exists := cs.ChildTarget(block.ParentID)
	if !exists {
		panic(fmt.Sprint("ConsensusSet is missing target of known block", block.ParentID))
	}

	// update fields
	bf.BlockID = block.ID()
	bf.Height++
	bf.Difficulty = target.Difficulty()
	bf.Target = target
	bf.Timestamp = block.Timestamp
	bf.TotalCoins = types.CalculateNumTurtleDexcoins(bf.Height)

	// calculate maturity timestamp
	var maturityTimestamp types.Timestamp
	if bf.Height > types.MaturityDelay {
		oldBlock, exists := cs.BlockAtHeight(bf.Height - types.MaturityDelay)
		if !exists {
			panic(fmt.Sprint("ConsensusSet is missing block at height", bf.Height-types.MaturityDelay))
		}
		maturityTimestamp = oldBlock.Timestamp
	}
	bf.MaturityTimestamp = maturityTimestamp

	// calculate hashrate by averaging last 'hashrateEstimationBlocks' blocks
	var estimatedHashrate types.Currency
	if bf.Height > hashrateEstimationBlocks {
		var totalDifficulty = bf.Target
		var oldestTimestamp types.Timestamp
		for i := types.BlockHeight(1); i < hashrateEstimationBlocks; i++ {
			b, exists := cs.BlockAtHeight(bf.Height - i)
			if !exists {
				panic(fmt.Sprint("ConsensusSet is missing block at height", bf.Height-hashrateEstimationBlocks))
			}
			target, exists := cs.ChildTarget(b.ParentID)
			if !exists {
				panic(fmt.Sprint("ConsensusSet is missing target of known block", b.ParentID))
			}
			totalDifficulty = totalDifficulty.AddDifficulties(target)
			oldestTimestamp = b.Timestamp
		}
		secondsPassed := bf.Timestamp - oldestTimestamp
		estimatedHashrate = totalDifficulty.Difficulty().Div64(uint64(secondsPassed))
	}
	bf.EstimatedHashrate = estimatedHashrate

	bf.MinerPayoutCount += uint64(len(block.MinerPayouts))
	bf.TransactionCount += uint64(len(block.Transactions))
	for _, txn := range block.Transactions {
		bf.TurtleDexcoinInputCount += uint64(len(txn.TurtleDexcoinInputs))
		bf.TurtleDexcoinOutputCount += uint64(len(txn.TurtleDexcoinOutputs))
		bf.FileContractCount += uint64(len(txn.FileContracts))
		bf.FileContractRevisionCount += uint64(len(txn.FileContractRevisions))
		bf.StorageProofCount += uint64(len(txn.StorageProofs))
		bf.TurtleDexfundInputCount += uint64(len(txn.TurtleDexfundInputs))
		bf.TurtleDexfundOutputCount += uint64(len(txn.TurtleDexfundOutputs))
		bf.MinerFeeCount += uint64(len(txn.MinerFees))
		bf.ArbitraryDataCount += uint64(len(txn.ArbitraryData))
		bf.TransactionSignatureCount += uint64(len(txn.TransactionSignatures))

		for _, fc := range txn.FileContracts {
			bf.TotalContractCost = bf.TotalContractCost.Add(fc.Payout)
			bf.TotalContractSize = bf.TotalContractSize.Add(types.NewCurrency64(fc.FileSize))
		}
		for _, fcr := range txn.FileContractRevisions {
			bf.TotalContractSize = bf.TotalContractSize.Add(types.NewCurrency64(fcr.NewFileSize))
			bf.TotalRevisionVolume = bf.TotalRevisionVolume.Add(types.NewCurrency64(fcr.NewFileSize))
		}
	}

	return bf
}

// Special handling for the genesis block. No other functions are called on it.
func dbAddGenesisBlock(tx *bolt.Tx) {
	id := types.GenesisID
	dbAddBlockID(tx, id, 0)

	// Add Genesis transactions to database
	for _, transaction := range types.GenesisBlock.Transactions {
		// Add Genesis Transaction to database
		txid := transaction.ID()
		dbAddTransactionID(tx, txid, 0)
		// Add Genesis TurtleDexcoin outputs to database
		for i, sco := range transaction.TurtleDexcoinOutputs {
			scoid := transaction.TurtleDexcoinOutputID(uint64(i))
			dbAddTurtleDexcoinOutputID(tx, scoid, txid)
			dbAddUnlockHash(tx, sco.UnlockHash, txid)
			dbAddTurtleDexcoinOutput(tx, scoid, sco)
		}

		// Add Geesis TurtleDexfund outputs to database
		for i, sfo := range transaction.TurtleDexfundOutputs {
			sfoid := transaction.TurtleDexfundOutputID(uint64(i))
			dbAddTurtleDexfundOutputID(tx, sfoid, txid)
			dbAddUnlockHash(tx, sfo.UnlockHash, txid)
			dbAddTurtleDexfundOutput(tx, sfoid, sfo)
		}
	}

	dbAddBlockFacts(tx, blockFacts{
		BlockFacts: modules.BlockFacts{
			BlockID:            id,
			Height:             0,
			Difficulty:         types.RootTarget.Difficulty(),
			Target:             types.RootTarget,
			TotalCoins:         types.CalculateCoinbase(0),
			TransactionCount:   uint64(len(types.GenesisBlock.Transactions)),
			TurtleDexcoinOutputCount: uint64(len(types.GenesisTurtleDexcoinAllocation)),
			TurtleDexfundOutputCount: uint64(len(types.GenesisTurtleDexfundAllocation)),
		},
		Timestamp: types.GenesisBlock.Timestamp,
	})
}
