package consensus

// applytransaction.go handles applying a transaction to the consensus set.
// There is an assumption that the transaction has already been verified.

import (
	"bytes"

	"github.com/turtledex/bolt"
	"github.com/turtledex/encoding"

	"github.com/turtledex/TurtleDexCore/build"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
)

// applyTurtleDexcoinInputs takes all of the ttdc inputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applyTurtleDexcoinInputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// Remove all ttdc inputs from the unspent ttdc outputs list.
	for _, sci := range t.TurtleDexcoinInputs {
		sco, err := getTurtleDexcoinOutput(tx, sci.ParentID)
		if build.DEBUG && err != nil {
			panic(err)
		}
		scod := modules.TurtleDexcoinOutputDiff{
			Direction:     modules.DiffRevert,
			ID:            sci.ParentID,
			TurtleDexcoinOutput: sco,
		}
		pb.TurtleDexcoinOutputDiffs = append(pb.TurtleDexcoinOutputDiffs, scod)
		commitTurtleDexcoinOutputDiff(tx, scod, modules.DiffApply)
	}
}

// applyTurtleDexcoinOutputs takes all of the ttdc outputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applyTurtleDexcoinOutputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// Add all ttdc outputs to the unspent ttdc outputs list.
	for i, sco := range t.TurtleDexcoinOutputs {
		scoid := t.TurtleDexcoinOutputID(uint64(i))
		scod := modules.TurtleDexcoinOutputDiff{
			Direction:     modules.DiffApply,
			ID:            scoid,
			TurtleDexcoinOutput: sco,
		}
		pb.TurtleDexcoinOutputDiffs = append(pb.TurtleDexcoinOutputDiffs, scod)
		commitTurtleDexcoinOutputDiff(tx, scod, modules.DiffApply)
	}
}

// applyFileContracts iterates through all of the file contracts in a
// transaction and applies them to the state, updating the diffs in the proccesed
// block.
func applyFileContracts(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for i, fc := range t.FileContracts {
		fcid := t.FileContractID(uint64(i))
		fcd := modules.FileContractDiff{
			Direction:    modules.DiffApply,
			ID:           fcid,
			FileContract: fc,
		}
		pb.FileContractDiffs = append(pb.FileContractDiffs, fcd)
		commitFileContractDiff(tx, fcd, modules.DiffApply)

		// Get the portion of the contract that goes into the siafund pool and
		// add it to the siafund pool.
		sfp := getTurtleDexfundPool(tx)
		sfpd := modules.TurtleDexfundPoolDiff{
			Direction: modules.DiffApply,
			Previous:  sfp,
			Adjusted:  sfp.Add(types.Tax(blockHeight(tx), fc.Payout)),
		}
		pb.TurtleDexfundPoolDiffs = append(pb.TurtleDexfundPoolDiffs, sfpd)
		commitTurtleDexfundPoolDiff(tx, sfpd, modules.DiffApply)
	}
}

// applyFileContractRevisions iterates through all of the file contract
// revisions in a transaction and applies them to the state, updating the diffs
// in the processed block.
func applyFileContractRevisions(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for _, fcr := range t.FileContractRevisions {
		fc, err := getFileContract(tx, fcr.ParentID)
		if build.DEBUG && err != nil {
			panic(err)
		}

		// Add the diff to delete the old file contract.
		fcd := modules.FileContractDiff{
			Direction:    modules.DiffRevert,
			ID:           fcr.ParentID,
			FileContract: fc,
		}
		pb.FileContractDiffs = append(pb.FileContractDiffs, fcd)
		commitFileContractDiff(tx, fcd, modules.DiffApply)

		// Add the diff to add the revised file contract.
		newFC := types.FileContract{
			FileSize:           fcr.NewFileSize,
			FileMerkleRoot:     fcr.NewFileMerkleRoot,
			WindowStart:        fcr.NewWindowStart,
			WindowEnd:          fcr.NewWindowEnd,
			Payout:             fc.Payout,
			ValidProofOutputs:  fcr.NewValidProofOutputs,
			MissedProofOutputs: fcr.NewMissedProofOutputs,
			UnlockHash:         fcr.NewUnlockHash,
			RevisionNumber:     fcr.NewRevisionNumber,
		}
		fcd = modules.FileContractDiff{
			Direction:    modules.DiffApply,
			ID:           fcr.ParentID,
			FileContract: newFC,
		}
		pb.FileContractDiffs = append(pb.FileContractDiffs, fcd)
		commitFileContractDiff(tx, fcd, modules.DiffApply)
	}
}

// applyTxStorageProofs iterates through all of the storage proofs in a
// transaction and applies them to the state, updating the diffs in the processed
// block.
func applyStorageProofs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for _, sp := range t.StorageProofs {
		fc, err := getFileContract(tx, sp.ParentID)
		if build.DEBUG && err != nil {
			panic(err)
		}

		// Add all of the outputs in the ValidProofOutputs of the contract.
		for i, vpo := range fc.ValidProofOutputs {
			spoid := sp.ParentID.StorageProofOutputID(types.ProofValid, uint64(i))
			dscod := modules.DelayedTurtleDexcoinOutputDiff{
				Direction:      modules.DiffApply,
				ID:             spoid,
				TurtleDexcoinOutput:  vpo,
				MaturityHeight: pb.Height + types.MaturityDelay,
			}
			pb.DelayedTurtleDexcoinOutputDiffs = append(pb.DelayedTurtleDexcoinOutputDiffs, dscod)
			commitDelayedTurtleDexcoinOutputDiff(tx, dscod, modules.DiffApply)
		}

		fcd := modules.FileContractDiff{
			Direction:    modules.DiffRevert,
			ID:           sp.ParentID,
			FileContract: fc,
		}
		pb.FileContractDiffs = append(pb.FileContractDiffs, fcd)
		commitFileContractDiff(tx, fcd, modules.DiffApply)
	}
}

// applyTxTurtleDexfundInputs takes all of the siafund inputs in a transaction and
// applies them to the state, updating the diffs in the processed block.
func applyTurtleDexfundInputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for _, sfi := range t.TurtleDexfundInputs {
		// Calculate the volume of ttdcs to put in the claim output.
		sfo, err := getTurtleDexfundOutput(tx, sfi.ParentID)
		if build.DEBUG && err != nil {
			panic(err)
		}
		claimPortion := getTurtleDexfundPool(tx).Sub(sfo.ClaimStart).Div(types.TurtleDexfundCount).Mul(sfo.Value)

		// Add the claim output to the delayed set of outputs.
		sco := types.TurtleDexcoinOutput{
			Value:      claimPortion,
			UnlockHash: sfi.ClaimUnlockHash,
		}
		sfoid := sfi.ParentID.TurtleDexClaimOutputID()
		dscod := modules.DelayedTurtleDexcoinOutputDiff{
			Direction:      modules.DiffApply,
			ID:             sfoid,
			TurtleDexcoinOutput:  sco,
			MaturityHeight: pb.Height + types.MaturityDelay,
		}
		pb.DelayedTurtleDexcoinOutputDiffs = append(pb.DelayedTurtleDexcoinOutputDiffs, dscod)
		commitDelayedTurtleDexcoinOutputDiff(tx, dscod, modules.DiffApply)

		// Create the siafund output diff and remove the output from the
		// consensus set.
		sfod := modules.TurtleDexfundOutputDiff{
			Direction:     modules.DiffRevert,
			ID:            sfi.ParentID,
			TurtleDexfundOutput: sfo,
		}
		pb.TurtleDexfundOutputDiffs = append(pb.TurtleDexfundOutputDiffs, sfod)
		commitTurtleDexfundOutputDiff(tx, sfod, modules.DiffApply)
	}
}

// applyTurtleDexfundOutputs applies a siafund output to the consensus set.
func applyTurtleDexfundOutputs(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	for i, sfo := range t.TurtleDexfundOutputs {
		sfoid := t.TurtleDexfundOutputID(uint64(i))
		sfo.ClaimStart = getTurtleDexfundPool(tx)
		sfod := modules.TurtleDexfundOutputDiff{
			Direction:     modules.DiffApply,
			ID:            sfoid,
			TurtleDexfundOutput: sfo,
		}
		pb.TurtleDexfundOutputDiffs = append(pb.TurtleDexfundOutputDiffs, sfod)
		commitTurtleDexfundOutputDiff(tx, sfod, modules.DiffApply)
	}
}

// applyArbitraryData applies arbitrary data to the consensus set. ArbitraryData
// is a field of the Transaction type whose structure is not fixed. This means
// that, via hardfork, new types of transaction can be introduced with minimal
// breakage by updating consensus code to recognize and act upon values encoded
// within the ArbitraryData field.
//
// Accordingly, this function dispatches on the various ArbitraryData values
// that are recognized by consensus. Currently, types.FoundationUnlockHashUpdate
// is the only recognized value.
func applyArbitraryData(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	// No ArbitraryData values were recognized prior to the Foundation hardfork.
	if pb.Height < types.FoundationHardforkHeight {
		return
	}
	for _, arb := range t.ArbitraryData {
		if bytes.HasPrefix(arb, types.SpecifierFoundation[:]) {
			var update types.FoundationUnlockHashUpdate
			err := encoding.Unmarshal(arb[types.SpecifierLen:], &update)
			if build.DEBUG && err != nil {
				// (Transaction).StandaloneValid ensures that decoding will not fail
				panic(err)
			}
			// Apply the update. First, save a copy of the old (i.e. current)
			// unlock hashes, so that we can revert later. Then set the new
			// unlock hashes.
			//
			// Importantly, we must only do this once per block; otherwise, for
			// complicated reasons involving diffs, we would not be able to
			// revert updates safely. So if we see that a copy has already been
			// recorded, we simply ignore the update; i.e. only the first update
			// in a block will be applied.
			if tx.Bucket(FoundationUnlockHashes).Get(encoding.Marshal(pb.Height)) != nil {
				continue
			}
			setPriorFoundationUnlockHashes(tx, pb.Height)
			setFoundationUnlockHashes(tx, update.NewPrimary, update.NewFailsafe)
			transferFoundationOutputs(tx, pb.Height, update.NewPrimary)
		}
	}
}

// transferFoundationOutputs transfers all unspent subsidy outputs to
// newPrimary. This allows subsidies to be recovered in the event that the
// primary key is lost or unusable when a subsidy is created.
func transferFoundationOutputs(tx *bolt.Tx, currentHeight types.BlockHeight, newPrimary types.UnlockHash) {
	for height := types.FoundationHardforkHeight; height < currentHeight; height += types.FoundationSubsidyFrequency {
		blockID, err := getPath(tx, height)
		if err != nil {
			if build.DEBUG {
				panic(err)
			}
			continue
		}
		id := blockID.FoundationSubsidyID()
		sco, err := getTurtleDexcoinOutput(tx, id)
		if err != nil {
			continue // output has already been spent
		}
		sco.UnlockHash = newPrimary
		removeTurtleDexcoinOutput(tx, id)
		addTurtleDexcoinOutput(tx, id, sco)
	}
}

// applyTransaction applies the contents of a transaction to the ConsensusSet.
// This produces a set of diffs, which are stored in the blockNode containing
// the transaction. No verification is done by this function.
func applyTransaction(tx *bolt.Tx, pb *processedBlock, t types.Transaction) {
	applyTurtleDexcoinInputs(tx, pb, t)
	applyTurtleDexcoinOutputs(tx, pb, t)
	applyFileContracts(tx, pb, t)
	applyFileContractRevisions(tx, pb, t)
	applyStorageProofs(tx, pb, t)
	applyTurtleDexfundInputs(tx, pb, t)
	applyTurtleDexfundOutputs(tx, pb, t)
	applyArbitraryData(tx, pb, t)
}
