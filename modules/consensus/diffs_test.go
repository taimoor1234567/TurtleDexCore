package consensus

import (
	"testing"

	"github.com/turtledex/bolt"

	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
)

// TestCommitDelayedTurtleDexcoinOutputDiffBadMaturity commits a delayed ttdc
// output that has a bad maturity height and triggers a panic.
func TestCommitDelayedTurtleDexcoinOutputDiffBadMaturity(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cst.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Trigger an inconsistency check.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expecting error after corrupting database")
		}
	}()

	// Commit a delayed ttdc output with maturity height = cs.height()+1
	maturityHeight := cst.cs.dbBlockHeight() - 1
	id := types.TurtleDexcoinOutputID{'1'}
	dsco := types.TurtleDexcoinOutput{Value: types.NewCurrency64(1)}
	dscod := modules.DelayedTurtleDexcoinOutputDiff{
		Direction:      modules.DiffApply,
		ID:             id,
		TurtleDexcoinOutput:  dsco,
		MaturityHeight: maturityHeight,
	}
	_ = cst.cs.db.Update(func(tx *bolt.Tx) error {
		commitDelayedTurtleDexcoinOutputDiff(tx, dscod, modules.DiffApply)
		return nil
	})
}

// TestCommitNodeDiffs probes the commitNodeDiffs method of the consensus set.
/*
func TestCommitNodeDiffs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
  	if err := cst.Close(); err != nil {
  		t.Fatal(err)
 	}
}()
	pb := cst.cs.dbCurrentProcessedBlock()
	_ = cst.cs.db.Update(func(tx *bolt.Tx) error {
		commitDiffSet(tx, pb, modules.DiffRevert) // pull the block node out of the consensus set.
		return nil
	})

	// For diffs that can be destroyed in the same block they are created,
	// create diffs that do just that. This has in the past caused issues upon
	// rewinding.
	scoid := types.TurtleDexcoinOutputID{'1'}
	scod0 := modules.TurtleDexcoinOutputDiff{
		Direction: modules.DiffApply,
		ID:        scoid,
	}
	scod1 := modules.TurtleDexcoinOutputDiff{
		Direction: modules.DiffRevert,
		ID:        scoid,
	}
	fcid := types.FileContractID{'2'}
	fcd0 := modules.FileContractDiff{
		Direction: modules.DiffApply,
		ID:        fcid,
	}
	fcd1 := modules.FileContractDiff{
		Direction: modules.DiffRevert,
		ID:        fcid,
	}
	sfoid := types.TurtleDexfundOutputID{'3'}
	sfod0 := modules.TurtleDexfundOutputDiff{
		Direction: modules.DiffApply,
		ID:        sfoid,
	}
	sfod1 := modules.TurtleDexfundOutputDiff{
		Direction: modules.DiffRevert,
		ID:        sfoid,
	}
	dscoid := types.TurtleDexcoinOutputID{'4'}
	dscod := modules.DelayedTurtleDexcoinOutputDiff{
		Direction:      modules.DiffApply,
		ID:             dscoid,
		MaturityHeight: cst.cs.dbBlockHeight() + types.MaturityDelay,
	}
	var siafundPool types.Currency
	err = cst.cs.db.Update(func(tx *bolt.Tx) error {
		siafundPool = getTurtleDexfundPool(tx)
		return nil
	})
	if err != nil {
		panic(err)
	}
	sfpd := modules.TurtleDexfundPoolDiff{
		Direction: modules.DiffApply,
		Previous:  siafundPool,
		Adjusted:  siafundPool.Add(types.NewCurrency64(1)),
	}
	pb.TurtleDexcoinOutputDiffs = append(pb.TurtleDexcoinOutputDiffs, scod0)
	pb.TurtleDexcoinOutputDiffs = append(pb.TurtleDexcoinOutputDiffs, scod1)
	pb.FileContractDiffs = append(pb.FileContractDiffs, fcd0)
	pb.FileContractDiffs = append(pb.FileContractDiffs, fcd1)
	pb.TurtleDexfundOutputDiffs = append(pb.TurtleDexfundOutputDiffs, sfod0)
	pb.TurtleDexfundOutputDiffs = append(pb.TurtleDexfundOutputDiffs, sfod1)
	pb.DelayedTurtleDexcoinOutputDiffs = append(pb.DelayedTurtleDexcoinOutputDiffs, dscod)
	pb.TurtleDexfundPoolDiffs = append(pb.TurtleDexfundPoolDiffs, sfpd)
	_ = cst.cs.db.Update(func(tx *bolt.Tx) error {
		createUpcomingDelayedOutputMaps(tx, pb, modules.DiffApply)
		return nil
	})
	_ = cst.cs.db.Update(func(tx *bolt.Tx) error {
		commitNodeDiffs(tx, pb, modules.DiffApply)
		return nil
	})
	exists := cst.cs.db.inTurtleDexcoinOutputs(scoid)
	if exists {
		t.Error("intradependent outputs not treated correctly")
	}
	exists = cst.cs.db.inFileContracts(fcid)
	if exists {
		t.Error("intradependent outputs not treated correctly")
	}
	exists = cst.cs.db.inTurtleDexfundOutputs(sfoid)
	if exists {
		t.Error("intradependent outputs not treated correctly")
	}
	_ = cst.cs.db.Update(func(tx *bolt.Tx) error {
		commitNodeDiffs(tx, pb, modules.DiffRevert)
		return nil
	})
	exists = cst.cs.db.inTurtleDexcoinOutputs(scoid)
	if exists {
		t.Error("intradependent outputs not treated correctly")
	}
	exists = cst.cs.db.inFileContracts(fcid)
	if exists {
		t.Error("intradependent outputs not treated correctly")
	}
	exists = cst.cs.db.inTurtleDexfundOutputs(sfoid)
	if exists {
		t.Error("intradependent outputs not treated correctly")
	}
}
*/

/*
// TestTurtleDexcoinOutputDiff applies and reverts a ttdc output diff, then
// triggers an inconsistency panic.
func TestCommitTurtleDexcoinOutputDiff(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
  	if err := cst.Close(); err != nil {
  		t.Fatal(err)
 	}
}()

	// Commit a ttdc output diff.
	initialScosLen := cst.cs.db.lenTurtleDexcoinOutputs()
	id := types.TurtleDexcoinOutputID{'1'}
	sco := types.TurtleDexcoinOutput{Value: types.NewCurrency64(1)}
	scod := modules.TurtleDexcoinOutputDiff{
		Direction:     modules.DiffApply,
		ID:            id,
		TurtleDexcoinOutput: sco,
	}
	cst.cs.commitTurtleDexcoinOutputDiff(scod, modules.DiffApply)
	if cst.cs.db.lenTurtleDexcoinOutputs() != initialScosLen+1 {
		t.Error("ttdc output diff set did not increase in size")
	}
	if cst.cs.db.getTurtleDexcoinOutputs(id).Value.Cmp(sco.Value) != 0 {
		t.Error("wrong ttdc output value after committing a diff")
	}

	// Rewind the diff.
	cst.cs.commitTurtleDexcoinOutputDiff(scod, modules.DiffRevert)
	if cst.cs.db.lenTurtleDexcoinOutputs() != initialScosLen {
		t.Error("ttdc output diff set did not increase in size")
	}
	exists := cst.cs.db.inTurtleDexcoinOutputs(id)
	if exists {
		t.Error("ttdc output was not reverted")
	}

	// Restore the diff and then apply the inverse diff.
	cst.cs.commitTurtleDexcoinOutputDiff(scod, modules.DiffApply)
	scod.Direction = modules.DiffRevert
	cst.cs.commitTurtleDexcoinOutputDiff(scod, modules.DiffApply)
	if cst.cs.db.lenTurtleDexcoinOutputs() != initialScosLen {
		t.Error("ttdc output diff set did not increase in size")
	}
	exists = cst.cs.db.inTurtleDexcoinOutputs(id)
	if exists {
		t.Error("ttdc output was not reverted")
	}

	// Revert the inverse diff.
	cst.cs.commitTurtleDexcoinOutputDiff(scod, modules.DiffRevert)
	if cst.cs.db.lenTurtleDexcoinOutputs() != initialScosLen+1 {
		t.Error("ttdc output diff set did not increase in size")
	}
	if cst.cs.db.getTurtleDexcoinOutputs(id).Value.Cmp(sco.Value) != 0 {
		t.Error("wrong ttdc output value after committing a diff")
	}

	// Trigger an inconsistency check.
	defer func() {
		r := recover()
		if r != errBadCommitTurtleDexcoinOutputDiff {
			t.Error("expecting errBadCommitTurtleDexcoinOutputDiff, got", r)
		}
	}()
	// Try reverting a revert diff that was already reverted. (add an object
	// that already exists)
	cst.cs.commitTurtleDexcoinOutputDiff(scod, modules.DiffRevert)
}
*/

/*
// TestCommitFileContracttDiff applies and reverts a file contract diff, then
// triggers an inconsistency panic.
func TestCommitFileContractDiff(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Commit a file contract diff.
	initialFcsLen := cst.cs.db.lenFileContracts()
	id := types.FileContractID{'1'}
	fc := types.FileContract{Payout: types.NewCurrency64(1)}
	fcd := modules.FileContractDiff{
		Direction:    modules.DiffApply,
		ID:           id,
		FileContract: fc,
	}
	cst.cs.commitFileContractDiff(fcd, modules.DiffApply)
	if cst.cs.db.lenFileContracts() != initialFcsLen+1 {
		t.Error("ttdc output diff set did not increase in size")
	}
	if cst.cs.db.getFileContracts(id).Payout.Cmp(fc.Payout) != 0 {
		t.Error("wrong ttdc output value after committing a diff")
	}

	// Rewind the diff.
	cst.cs.commitFileContractDiff(fcd, modules.DiffRevert)
	if cst.cs.db.lenFileContracts() != initialFcsLen {
		t.Error("ttdc output diff set did not increase in size")
	}
	exists := cst.cs.db.inFileContracts(id)
	if exists {
		t.Error("ttdc output was not reverted")
	}

	// Restore the diff and then apply the inverse diff.
	cst.cs.commitFileContractDiff(fcd, modules.DiffApply)
	fcd.Direction = modules.DiffRevert
	cst.cs.commitFileContractDiff(fcd, modules.DiffApply)
	if cst.cs.db.lenFileContracts() != initialFcsLen {
		t.Error("ttdc output diff set did not increase in size")
	}
	exists = cst.cs.db.inFileContracts(id)
	if exists {
		t.Error("ttdc output was not reverted")
	}

	// Revert the inverse diff.
	cst.cs.commitFileContractDiff(fcd, modules.DiffRevert)
	if cst.cs.db.lenFileContracts() != initialFcsLen+1 {
		t.Error("ttdc output diff set did not increase in size")
	}
	if cst.cs.db.getFileContracts(id).Payout.Cmp(fc.Payout) != 0 {
		t.Error("wrong ttdc output value after committing a diff")
	}

	// Trigger an inconsistency check.
	defer func() {
		r := recover()
		if r != errBadCommitFileContractDiff {
			t.Error("expecting errBadCommitFileContractDiff, got", r)
		}
	}()
	// Try reverting an apply diff that was already reverted. (remove an object
	// that was already removed)
	fcd.Direction = modules.DiffApply                      // Object currently exists, but make the direction 'apply'.
	cst.cs.commitFileContractDiff(fcd, modules.DiffRevert) // revert the application.
	cst.cs.commitFileContractDiff(fcd, modules.DiffRevert) // revert the application again, in error.
}
*/

// TestTurtleDexfundOutputDiff applies and reverts a siafund output diff, then
// triggers an inconsistency panic.
/*
func TestCommitTurtleDexfundOutputDiff(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Commit a siafund output diff.
	initialScosLen := cst.cs.db.lenTurtleDexfundOutputs()
	id := types.TurtleDexfundOutputID{'1'}
	sfo := types.TurtleDexfundOutput{Value: types.NewCurrency64(1)}
	sfod := modules.TurtleDexfundOutputDiff{
		Direction:     modules.DiffApply,
		ID:            id,
		TurtleDexfundOutput: sfo,
	}
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffApply)
	if cst.cs.db.lenTurtleDexfundOutputs() != initialScosLen+1 {
		t.Error("siafund output diff set did not increase in size")
	}
	sfo1 := cst.cs.db.getTurtleDexfundOutputs(id)
	if sfo1.Value.Cmp(sfo.Value) != 0 {
		t.Error("wrong siafund output value after committing a diff")
	}

	// Rewind the diff.
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffRevert)
	if cst.cs.db.lenTurtleDexfundOutputs() != initialScosLen {
		t.Error("siafund output diff set did not increase in size")
	}
	exists := cst.cs.db.inTurtleDexfundOutputs(id)
	if exists {
		t.Error("siafund output was not reverted")
	}

	// Restore the diff and then apply the inverse diff.
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffApply)
	sfod.Direction = modules.DiffRevert
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffApply)
	if cst.cs.db.lenTurtleDexfundOutputs() != initialScosLen {
		t.Error("siafund output diff set did not increase in size")
	}
	exists = cst.cs.db.inTurtleDexfundOutputs(id)
	if exists {
		t.Error("siafund output was not reverted")
	}

	// Revert the inverse diff.
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffRevert)
	if cst.cs.db.lenTurtleDexfundOutputs() != initialScosLen+1 {
		t.Error("siafund output diff set did not increase in size")
	}
	sfo2 := cst.cs.db.getTurtleDexfundOutputs(id)
	if sfo2.Value.Cmp(sfo.Value) != 0 {
		t.Error("wrong siafund output value after committing a diff")
	}

	// Trigger an inconsistency check.
	defer func() {
		r := recover()
		if r != errBadCommitTurtleDexfundOutputDiff {
			t.Error("expecting errBadCommitTurtleDexfundOutputDiff, got", r)
		}
	}()
	// Try applying a revert diff that was already applied. (remove an object
	// that was already removed)
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffApply) // Remove the object.
	cst.cs.commitTurtleDexfundOutputDiff(sfod, modules.DiffApply) // Remove the object again.
}
*/

// TestCommitDelayedTurtleDexcoinOutputDiff probes the commitDelayedTurtleDexcoinOutputDiff
// method of the consensus set.
/*
func TestCommitDelayedTurtleDexcoinOutputDiff(t *testing.T) {
	t.Skip("test isn't working, but checks the wrong code anyway")
	if testing.Short() {
		t.Skip()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Commit a delayed ttdc output with maturity height = cs.height()+1
	maturityHeight := cst.cs.height() + 1
	initialDscosLen := cst.cs.db.lenDelayedTurtleDexcoinOutputsHeight(maturityHeight)
	id := types.TurtleDexcoinOutputID{'1'}
	dsco := types.TurtleDexcoinOutput{Value: types.NewCurrency64(1)}
	dscod := modules.DelayedTurtleDexcoinOutputDiff{
		Direction:      modules.DiffApply,
		ID:             id,
		TurtleDexcoinOutput:  dsco,
		MaturityHeight: maturityHeight,
	}
	cst.cs.commitDelayedTurtleDexcoinOutputDiff(dscod, modules.DiffApply)
	if cst.cs.db.lenDelayedTurtleDexcoinOutputsHeight(maturityHeight) != initialDscosLen+1 {
		t.Fatal("delayed output diff set did not increase in size")
	}
	if cst.cs.db.getDelayedTurtleDexcoinOutputs(maturityHeight, id).Value.Cmp(dsco.Value) != 0 {
		t.Error("wrong delayed ttdc output value after committing a diff")
	}

	// Rewind the diff.
	cst.cs.commitDelayedTurtleDexcoinOutputDiff(dscod, modules.DiffRevert)
	if cst.cs.db.lenDelayedTurtleDexcoinOutputsHeight(maturityHeight) != initialDscosLen {
		t.Error("ttdc output diff set did not increase in size")
	}
	exists := cst.cs.db.inDelayedTurtleDexcoinOutputsHeight(maturityHeight, id)
	if exists {
		t.Error("ttdc output was not reverted")
	}

	// Restore the diff and then apply the inverse diff.
	cst.cs.commitDelayedTurtleDexcoinOutputDiff(dscod, modules.DiffApply)
	dscod.Direction = modules.DiffRevert
	cst.cs.commitDelayedTurtleDexcoinOutputDiff(dscod, modules.DiffApply)
	if cst.cs.db.lenDelayedTurtleDexcoinOutputsHeight(maturityHeight) != initialDscosLen {
		t.Error("ttdc output diff set did not increase in size")
	}
	exists = cst.cs.db.inDelayedTurtleDexcoinOutputsHeight(maturityHeight, id)
	if exists {
		t.Error("ttdc output was not reverted")
	}

	// Revert the inverse diff.
	cst.cs.commitDelayedTurtleDexcoinOutputDiff(dscod, modules.DiffRevert)
	if cst.cs.db.lenDelayedTurtleDexcoinOutputsHeight(maturityHeight) != initialDscosLen+1 {
		t.Error("ttdc output diff set did not increase in size")
	}
	if cst.cs.db.getDelayedTurtleDexcoinOutputs(maturityHeight, id).Value.Cmp(dsco.Value) != 0 {
		t.Error("wrong ttdc output value after committing a diff")
	}

	// Trigger an inconsistency check.
	defer func() {
		r := recover()
		if r != errBadCommitDelayedTurtleDexcoinOutputDiff {
			t.Error("expecting errBadCommitDelayedTurtleDexcoinOutputDiff, got", r)
		}
	}()
	// Try applying an apply diff that was already applied. (add an object
	// that already exists)
	dscod.Direction = modules.DiffApply                             // set the direction to apply
	cst.cs.commitDelayedTurtleDexcoinOutputDiff(dscod, modules.DiffApply) // apply an already existing delayed output.
}
*/

/*
// TestCommitTurtleDexfundPoolDiff probes the commitTurtleDexfundPoolDiff method of the
// consensus set.
func TestCommitTurtleDexfundPoolDiff(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Apply two siafund pool diffs, and then a diff with 0 change. Then revert
	// them all.
	initial := cst.cs.siafundPool
	adjusted1 := initial.Add(types.NewCurrency64(200))
	adjusted2 := adjusted1.Add(types.NewCurrency64(500))
	adjusted3 := adjusted2.Add(types.NewCurrency64(0))
	sfpd1 := modules.TurtleDexfundPoolDiff{
		Direction: modules.DiffApply,
		Previous:  initial,
		Adjusted:  adjusted1,
	}
	sfpd2 := modules.TurtleDexfundPoolDiff{
		Direction: modules.DiffApply,
		Previous:  adjusted1,
		Adjusted:  adjusted2,
	}
	sfpd3 := modules.TurtleDexfundPoolDiff{
		Direction: modules.DiffApply,
		Previous:  adjusted2,
		Adjusted:  adjusted3,
	}
	cst.cs.commitTurtleDexfundPoolDiff(sfpd1, modules.DiffApply)
	if cst.cs.siafundPool.Cmp(adjusted1) != 0 {
		t.Error("siafund pool was not adjusted correctly")
	}
	cst.cs.commitTurtleDexfundPoolDiff(sfpd2, modules.DiffApply)
	if cst.cs.siafundPool.Cmp(adjusted2) != 0 {
		t.Error("second siafund pool adjustment was flawed")
	}
	cst.cs.commitTurtleDexfundPoolDiff(sfpd3, modules.DiffApply)
	if cst.cs.siafundPool.Cmp(adjusted3) != 0 {
		t.Error("second siafund pool adjustment was flawed")
	}
	cst.cs.commitTurtleDexfundPoolDiff(sfpd3, modules.DiffRevert)
	if cst.cs.siafundPool.Cmp(adjusted2) != 0 {
		t.Error("reverting second adjustment was flawed")
	}
	cst.cs.commitTurtleDexfundPoolDiff(sfpd2, modules.DiffRevert)
	if cst.cs.siafundPool.Cmp(adjusted1) != 0 {
		t.Error("reverting second adjustment was flawed")
	}
	cst.cs.commitTurtleDexfundPoolDiff(sfpd1, modules.DiffRevert)
	if cst.cs.siafundPool.Cmp(initial) != 0 {
		t.Error("reverting first adjustment was flawed")
	}

	// Do a chaining set of panics. First apply a negative pool adjustment,
	// then revert the pool diffs in the wrong order, than apply the pool diffs
	// in the wrong order.
	defer func() {
		r := recover()
		if r != errApplyTurtleDexfundPoolDiffMismatch {
			t.Error("expecting errApplyTurtleDexfundPoolDiffMismatch, got", r)
		}
	}()
	defer func() {
		r := recover()
		if r != errRevertTurtleDexfundPoolDiffMismatch {
			t.Error("expecting errRevertTurtleDexfundPoolDiffMismatch, got", r)
		}
		cst.cs.commitTurtleDexfundPoolDiff(sfpd1, modules.DiffApply)
	}()
	defer func() {
		r := recover()
		if r != errNonApplyTurtleDexfundPoolDiff {
			t.Error(r)
		}
		cst.cs.commitTurtleDexfundPoolDiff(sfpd1, modules.DiffRevert)
	}()
	defer func() {
		r := recover()
		if r != errNegativePoolAdjustment {
			t.Error("expecting errNegativePoolAdjustment, got", r)
		}
		sfpd2.Direction = modules.DiffRevert
		cst.cs.commitTurtleDexfundPoolDiff(sfpd2, modules.DiffApply)
	}()
	cst.cs.commitTurtleDexfundPoolDiff(sfpd1, modules.DiffApply)
	cst.cs.commitTurtleDexfundPoolDiff(sfpd2, modules.DiffApply)
	negativeAdjustment := adjusted2.Sub(types.NewCurrency64(100))
	negativeSfpd := modules.TurtleDexfundPoolDiff{
		Previous: adjusted3,
		Adjusted: negativeAdjustment,
	}
	cst.cs.commitTurtleDexfundPoolDiff(negativeSfpd, modules.DiffApply)
}
*/

/*
// TestDeleteObsoleteDelayedOutputMapsSanity probes the sanity checks of the
// deleteObsoleteDelayedOutputMaps method of the consensus set.
func TestDeleteObsoleteDelayedOutputMapsSanity(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	pb := cst.cs.currentProcessedBlock()
	err = cst.cs.db.Update(func(tx *bolt.Tx) error {
		return commitDiffSet(tx, pb, modules.DiffRevert)
	})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expecting an error after corrupting the database")
		}
	}()
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expecting an error after corrupting the database")
		}

		// Trigger a panic by deleting a map with outputs in it during revert.
		err = cst.cs.db.Update(func(tx *bolt.Tx) error {
			return createUpcomingDelayedOutputMaps(tx, pb, modules.DiffApply)
		})
		if err != nil {
			t.Fatal(err)
		}
		err = cst.cs.db.Update(func(tx *bolt.Tx) error {
			return commitNodeDiffs(tx, pb, modules.DiffApply)
		})
		if err != nil {
			t.Fatal(err)
		}
		err = cst.cs.db.Update(func(tx *bolt.Tx) error {
			return deleteObsoleteDelayedOutputMaps(tx, pb, modules.DiffRevert)
		})
		if err != nil {
			t.Fatal(err)
		}
	}()

	// Trigger a panic by deleting a map with outputs in it during apply.
	err = cst.cs.db.Update(func(tx *bolt.Tx) error {
		return deleteObsoleteDelayedOutputMaps(tx, pb, modules.DiffApply)
	})
	if err != nil {
		t.Fatal(err)
	}
}
*/

/*
// TestGenerateAndApplyDiffSanity triggers the sanity checks in the
// generateAndApplyDiff method of the consensus set.
func TestGenerateAndApplyDiffSanity(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	pb := cst.cs.currentProcessedBlock()
	cst.cs.commitDiffSet(pb, modules.DiffRevert)

	defer func() {
		r := recover()
		if r != errRegenerateDiffs {
			t.Error("expected errRegenerateDiffs, got", r)
		}
	}()
	defer func() {
		r := recover()
		if r != errInvalidSuccessor {
			t.Error("expected errInvalidSuccessor, got", r)
		}

		// Trigger errRegenerteDiffs
		_ = cst.cs.generateAndApplyDiff(pb)
	}()

	// Trigger errInvalidSuccessor
	parent := cst.cs.db.getBlockMap(pb.Parent)
	parent.DiffsGenerated = false
	_ = cst.cs.generateAndApplyDiff(parent)
}
*/
