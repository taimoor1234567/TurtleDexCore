package consensus

import (
	"testing"

	"github.com/turtledex/TurtleDexCore/types"
	"github.com/turtledex/bolt"
	"github.com/turtledex/encoding"
)

/*
// TestApplyTurtleDexcoinInputs probes the applyTurtleDexcoinInputs method of the consensus
// set.
func TestApplyTurtleDexcoinInputs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a consensus set and get it to 3 ttdc outputs. The consensus
	// set starts with 2 ttdc outputs, mining a block will add another.
	cst, err := createConsensusSetTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cst.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	b, _ := cst.miner.FindBlock()
	err = cst.cs.AcceptBlock(b)
	if err != nil {
		t.Fatal(err)
	}

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Fetch the output id's of each ttdc output in the consensus set.
	var ids []types.TurtleDexcoinOutputID
	cst.cs.db.forEachTurtleDexcoinOutputs(func(id types.TurtleDexcoinOutputID, sco types.TurtleDexcoinOutput) {
		ids = append(ids, id)
	})

	// Apply a transaction with a single ttdc input.
	txn := types.Transaction{
		TurtleDexcoinInputs: []types.TurtleDexcoinInput{
			{ParentID: ids[0]},
		},
	}
	cst.cs.applyTurtleDexcoinInputs(pb, txn)
	exists := cst.cs.db.inTurtleDexcoinOutputs(ids[0])
	if exists {
		t.Error("Failed to conusme a ttdc output")
	}
	if cst.cs.db.lenTurtleDexcoinOutputs() != 2 {
		t.Error("ttdc outputs not correctly updated")
	}
	if len(pb.TurtleDexcoinOutputDiffs) != 1 {
		t.Error("block node was not updated for single transaction")
	}
	if pb.TurtleDexcoinOutputDiffs[0].Direction != modules.DiffRevert {
		t.Error("wrong diff direction applied when consuming a ttdc output")
	}
	if pb.TurtleDexcoinOutputDiffs[0].ID != ids[0] {
		t.Error("wrong id used when consuming a ttdc output")
	}

	// Apply a transaction with two ttdc inputs.
	txn = types.Transaction{
		TurtleDexcoinInputs: []types.TurtleDexcoinInput{
			{ParentID: ids[1]},
			{ParentID: ids[2]},
		},
	}
	cst.cs.applyTurtleDexcoinInputs(pb, txn)
	if cst.cs.db.lenTurtleDexcoinOutputs() != 0 {
		t.Error("failed to consume all ttdc outputs in the consensus set")
	}
	if len(pb.TurtleDexcoinOutputDiffs) != 3 {
		t.Error("processed block was not updated for single transaction")
	}
}

// TestMisuseApplyTurtleDexcoinInputs misuses applyTurtleDexcoinInput and checks that a
// panic was triggered.
func TestMisuseApplyTurtleDexcoinInputs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Fetch the output id's of each ttdc output in the consensus set.
	var ids []types.TurtleDexcoinOutputID
	cst.cs.db.forEachTurtleDexcoinOutputs(func(id types.TurtleDexcoinOutputID, sco types.TurtleDexcoinOutput) {
		ids = append(ids, id)
	})

	// Apply a transaction with a single ttdc input.
	txn := types.Transaction{
		TurtleDexcoinInputs: []types.TurtleDexcoinInput{
			{ParentID: ids[0]},
		},
	}
	cst.cs.applyTurtleDexcoinInputs(pb, txn)

	// Trigger the panic that occurs when an output is applied incorrectly, and
	// perform a catch to read the error that is created.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expecting error after corrupting database")
		}
	}()
	cst.cs.applyTurtleDexcoinInputs(pb, txn)
}

// TestApplyTurtleDexcoinOutputs probes the applyTurtleDexcoinOutput method of the
// consensus set.
func TestApplyTurtleDexcoinOutputs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with a single ttdc output.
	txn := types.Transaction{
		TurtleDexcoinOutputs: []types.TurtleDexcoinOutput{{}},
	}
	cst.cs.applyTurtleDexcoinOutputs(pb, txn)
	scoid := txn.TurtleDexcoinOutputID(0)
	exists := cst.cs.db.inTurtleDexcoinOutputs(scoid)
	if !exists {
		t.Error("Failed to create ttdc output")
	}
	if cst.cs.db.lenTurtleDexcoinOutputs() != 3 { // 3 because createConsensusSetTester has 2 initially.
		t.Error("ttdc outputs not correctly updated")
	}
	if len(pb.TurtleDexcoinOutputDiffs) != 1 {
		t.Error("block node was not updated for single element transaction")
	}
	if pb.TurtleDexcoinOutputDiffs[0].Direction != modules.DiffApply {
		t.Error("wrong diff direction applied when creating a ttdc output")
	}
	if pb.TurtleDexcoinOutputDiffs[0].ID != scoid {
		t.Error("wrong id used when creating a ttdc output")
	}

	// Apply a transaction with 2 ttdc outputs.
	txn = types.Transaction{
		TurtleDexcoinOutputs: []types.TurtleDexcoinOutput{
			{Value: types.NewCurrency64(1)},
			{Value: types.NewCurrency64(2)},
		},
	}
	cst.cs.applyTurtleDexcoinOutputs(pb, txn)
	scoid0 := txn.TurtleDexcoinOutputID(0)
	scoid1 := txn.TurtleDexcoinOutputID(1)
	exists = cst.cs.db.inTurtleDexcoinOutputs(scoid0)
	if !exists {
		t.Error("Failed to create ttdc output")
	}
	exists = cst.cs.db.inTurtleDexcoinOutputs(scoid1)
	if !exists {
		t.Error("Failed to create ttdc output")
	}
	if cst.cs.db.lenTurtleDexcoinOutputs() != 5 { // 5 because createConsensusSetTester has 2 initially.
		t.Error("ttdc outputs not correctly updated")
	}
	if len(pb.TurtleDexcoinOutputDiffs) != 3 {
		t.Error("block node was not updated correctly")
	}
}

// TestMisuseApplyTurtleDexcoinOutputs misuses applyTurtleDexcoinOutputs and checks that a
// panic was triggered.
func TestMisuseApplyTurtleDexcoinOutputs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with a single ttdc output.
	txn := types.Transaction{
		TurtleDexcoinOutputs: []types.TurtleDexcoinOutput{{}},
	}
	cst.cs.applyTurtleDexcoinOutputs(pb, txn)

	// Trigger the panic that occurs when an output is applied incorrectly, and
	// perform a catch to read the error that is created.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("no panic occurred when misusing applyTurtleDexcoinInput")
		}
	}()
	cst.cs.applyTurtleDexcoinOutputs(pb, txn)
}

// TestApplyFileContracts probes the applyFileContracts method of the
// consensus set.
func TestApplyFileContracts(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with a single file contract.
	txn := types.Transaction{
		FileContracts: []types.FileContract{{}},
	}
	cst.cs.applyFileContracts(pb, txn)
	fcid := txn.FileContractID(0)
	exists := cst.cs.db.inFileContracts(fcid)
	if !exists {
		t.Error("Failed to create file contract")
	}
	if cst.cs.db.lenFileContracts() != 1 {
		t.Error("file contracts not correctly updated")
	}
	if len(pb.FileContractDiffs) != 1 {
		t.Error("block node was not updated for single element transaction")
	}
	if pb.FileContractDiffs[0].Direction != modules.DiffApply {
		t.Error("wrong diff direction applied when creating a file contract")
	}
	if pb.FileContractDiffs[0].ID != fcid {
		t.Error("wrong id used when creating a file contract")
	}

	// Apply a transaction with 2 file contracts.
	txn = types.Transaction{
		FileContracts: []types.FileContract{
			{Payout: types.NewCurrency64(1)},
			{Payout: types.NewCurrency64(300e3)},
		},
	}
	cst.cs.applyFileContracts(pb, txn)
	fcid0 := txn.FileContractID(0)
	fcid1 := txn.FileContractID(1)
	exists = cst.cs.db.inFileContracts(fcid0)
	if !exists {
		t.Error("Failed to create file contract")
	}
	exists = cst.cs.db.inFileContracts(fcid1)
	if !exists {
		t.Error("Failed to create file contract")
	}
	if cst.cs.db.lenFileContracts() != 3 {
		t.Error("file contracts not correctly updated")
	}
	if len(pb.FileContractDiffs) != 3 {
		t.Error("block node was not updated correctly")
	}
	if cst.cs.siafundPool.Cmp64(10e3) != 0 {
		t.Error("siafund pool did not update correctly upon creation of a file contract")
	}
}

// TestMisuseApplyFileContracts misuses applyFileContracts and checks that a
// panic was triggered.
func TestMisuseApplyFileContracts(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with a single file contract.
	txn := types.Transaction{
		FileContracts: []types.FileContract{{}},
	}
	cst.cs.applyFileContracts(pb, txn)

	// Trigger the panic that occurs when an output is applied incorrectly, and
	// perform a catch to read the error that is created.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("no panic occurred when misusing applyTurtleDexcoinInput")
		}
	}()
	cst.cs.applyFileContracts(pb, txn)
}

// TestApplyFileContractRevisions probes the applyFileContractRevisions method
// of the consensus set.
func TestApplyFileContractRevisions(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with two file contracts - that way there is
	// something to revise.
	txn := types.Transaction{
		FileContracts: []types.FileContract{
			{},
			{Payout: types.NewCurrency64(1)},
		},
	}
	cst.cs.applyFileContracts(pb, txn)
	fcid0 := txn.FileContractID(0)
	fcid1 := txn.FileContractID(1)

	// Apply a single file contract revision.
	txn = types.Transaction{
		FileContractRevisions: []types.FileContractRevision{
			{
				ParentID:    fcid0,
				NewFileSize: 1,
			},
		},
	}
	cst.cs.applyFileContractRevisions(pb, txn)
	exists := cst.cs.db.inFileContracts(fcid0)
	if !exists {
		t.Error("Revision killed a file contract")
	}
	fc := cst.cs.db.getFileContracts(fcid0)
	if fc.FileSize != 1 {
		t.Error("file contract filesize not properly updated")
	}
	if cst.cs.db.lenFileContracts() != 2 {
		t.Error("file contracts not correctly updated")
	}
	if len(pb.FileContractDiffs) != 4 { // 2 creating the initial contracts, 1 to remove the old, 1 to add the revision.
		t.Error("block node was not updated for single element transaction")
	}
	if pb.FileContractDiffs[2].Direction != modules.DiffRevert {
		t.Error("wrong diff direction applied when revising a file contract")
	}
	if pb.FileContractDiffs[3].Direction != modules.DiffApply {
		t.Error("wrong diff direction applied when revising a file contract")
	}
	if pb.FileContractDiffs[2].ID != fcid0 {
		t.Error("wrong id used when revising a file contract")
	}
	if pb.FileContractDiffs[3].ID != fcid0 {
		t.Error("wrong id used when revising a file contract")
	}

	// Apply a transaction with 2 file contract revisions.
	txn = types.Transaction{
		FileContractRevisions: []types.FileContractRevision{
			{
				ParentID:    fcid0,
				NewFileSize: 2,
			},
			{
				ParentID:    fcid1,
				NewFileSize: 3,
			},
		},
	}
	cst.cs.applyFileContractRevisions(pb, txn)
	exists = cst.cs.db.inFileContracts(fcid0)
	if !exists {
		t.Error("Revision ate file contract")
	}
	fc0 := cst.cs.db.getFileContracts(fcid0)
	exists = cst.cs.db.inFileContracts(fcid1)
	if !exists {
		t.Error("Revision ate file contract")
	}
	fc1 := cst.cs.db.getFileContracts(fcid1)
	if fc0.FileSize != 2 {
		t.Error("Revision not correctly applied")
	}
	if fc1.FileSize != 3 {
		t.Error("Revision not correctly applied")
	}
	if cst.cs.db.lenFileContracts() != 2 {
		t.Error("file contracts not correctly updated")
	}
	if len(pb.FileContractDiffs) != 8 {
		t.Error("block node was not updated correctly")
	}
}

// TestMisuseApplyFileContractRevisions misuses applyFileContractRevisions and
// checks that a panic was triggered.
func TestMisuseApplyFileContractRevisions(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Trigger a panic from revising a nonexistent file contract.
	defer func() {
		r := recover()
		if r != errNilItem {
			t.Error("no panic occurred when misusing applyTurtleDexcoinInput")
		}
	}()
	txn := types.Transaction{
		FileContractRevisions: []types.FileContractRevision{{}},
	}
	cst.cs.applyFileContractRevisions(pb, txn)
}

// TestApplyStorageProofs probes the applyStorageProofs method of the consensus
// set.
func TestApplyStorageProofs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)
	pb.Height = cst.cs.height()

	// Apply a transaction with two file contracts - there is a reason to
	// create a storage proof.
	txn := types.Transaction{
		FileContracts: []types.FileContract{
			{
				Payout: types.NewCurrency64(300e3),
				ValidProofOutputs: []types.TurtleDexcoinOutput{
					{Value: types.NewCurrency64(290e3)},
				},
			},
			{},
			{
				Payout: types.NewCurrency64(600e3),
				ValidProofOutputs: []types.TurtleDexcoinOutput{
					{Value: types.NewCurrency64(280e3)},
					{Value: types.NewCurrency64(300e3)},
				},
			},
		},
	}
	cst.cs.applyFileContracts(pb, txn)
	fcid0 := txn.FileContractID(0)
	fcid1 := txn.FileContractID(1)
	fcid2 := txn.FileContractID(2)

	// Apply a single storage proof.
	txn = types.Transaction{
		StorageProofs: []types.StorageProof{{ParentID: fcid0}},
	}
	cst.cs.applyStorageProofs(pb, txn)
	exists := cst.cs.db.inFileContracts(fcid0)
	if exists {
		t.Error("Storage proof did not disable a file contract.")
	}
	if cst.cs.db.lenFileContracts() != 2 {
		t.Error("file contracts not correctly updated")
	}
	if len(pb.FileContractDiffs) != 4 { // 3 creating the initial contracts, 1 for the storage proof.
		t.Error("block node was not updated for single element transaction")
	}
	if pb.FileContractDiffs[3].Direction != modules.DiffRevert {
		t.Error("wrong diff direction applied when revising a file contract")
	}
	if pb.FileContractDiffs[3].ID != fcid0 {
		t.Error("wrong id used when revising a file contract")
	}
	spoid0 := fcid0.StorageProofOutputID(types.ProofValid, 0)
	exists = cst.cs.db.inDelayedTurtleDexcoinOutputsHeight(pb.Height+types.MaturityDelay, spoid0)
	if !exists {
		t.Error("storage proof output not created after applying a storage proof")
	}
	sco := cst.cs.db.getDelayedTurtleDexcoinOutputs(pb.Height+types.MaturityDelay, spoid0)
	if sco.Value.Cmp64(290e3) != 0 {
		t.Error("storage proof output was created with the wrong value")
	}

	// Apply a transaction with 2 storage proofs.
	txn = types.Transaction{
		StorageProofs: []types.StorageProof{
			{ParentID: fcid1},
			{ParentID: fcid2},
		},
	}
	cst.cs.applyStorageProofs(pb, txn)
	exists = cst.cs.db.inFileContracts(fcid1)
	if exists {
		t.Error("Storage proof failed to consume file contract.")
	}
	exists = cst.cs.db.inFileContracts(fcid2)
	if exists {
		t.Error("storage proof did not consume file contract")
	}
	if cst.cs.db.lenFileContracts() != 0 {
		t.Error("file contracts not correctly updated")
	}
	if len(pb.FileContractDiffs) != 6 {
		t.Error("block node was not updated correctly")
	}
	spoid1 := fcid1.StorageProofOutputID(types.ProofValid, 0)
	exists = cst.cs.db.inTurtleDexcoinOutputs(spoid1)
	if exists {
		t.Error("output created when file contract had no corresponding output")
	}
	spoid2 := fcid2.StorageProofOutputID(types.ProofValid, 0)
	exists = cst.cs.db.inDelayedTurtleDexcoinOutputsHeight(pb.Height+types.MaturityDelay, spoid2)
	if !exists {
		t.Error("no output created by first output of file contract")
	}
	sco = cst.cs.db.getDelayedTurtleDexcoinOutputs(pb.Height+types.MaturityDelay, spoid2)
	if sco.Value.Cmp64(280e3) != 0 {
		t.Error("first ttdc output created has wrong value")
	}
	spoid3 := fcid2.StorageProofOutputID(types.ProofValid, 1)
	exists = cst.cs.db.inDelayedTurtleDexcoinOutputsHeight(pb.Height+types.MaturityDelay, spoid3)
	if !exists {
		t.Error("second output not created for storage proof")
	}
	sco = cst.cs.db.getDelayedTurtleDexcoinOutputs(pb.Height+types.MaturityDelay, spoid3)
	if sco.Value.Cmp64(300e3) != 0 {
		t.Error("second ttdc output has wrong value")
	}
	if cst.cs.siafundPool.Cmp64(30e3) != 0 {
		t.Error("siafund pool not being added up correctly")
	}
}

// TestNonexistentStorageProof applies a storage proof which points to a
// nonextentent parent.
func TestNonexistentStorageProof(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Trigger a panic by applying a storage proof for a nonexistent file
	// contract.
	defer func() {
		r := recover()
		if r != errNilItem {
			t.Error("no panic occurred when misusing applyTurtleDexcoinInput")
		}
	}()
	txn := types.Transaction{
		StorageProofs: []types.StorageProof{{}},
	}
	cst.cs.applyStorageProofs(pb, txn)
}

// TestDuplicateStorageProof applies a storage proof which has already been
// applied.
func TestDuplicateStorageProof(t *testing.T) {
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

	// Create a block node.
	pb := new(processedBlock)
	pb.Height = cst.cs.height()

	// Create a file contract for the storage proof to prove.
	txn0 := types.Transaction{
		FileContracts: []types.FileContract{
			{
				Payout: types.NewCurrency64(300e3),
				ValidProofOutputs: []types.TurtleDexcoinOutput{
					{Value: types.NewCurrency64(290e3)},
				},
			},
		},
	}
	cst.cs.applyFileContracts(pb, txn0)
	fcid := txn0.FileContractID(0)

	// Apply a single storage proof.
	txn1 := types.Transaction{
		StorageProofs: []types.StorageProof{{ParentID: fcid}},
	}
	cst.cs.applyStorageProofs(pb, txn1)

	// Trigger a panic by applying the storage proof again.
	defer func() {
		r := recover()
		if r != ErrDuplicateValidProofOutput {
			t.Error("failed to trigger ErrDuplicateValidProofOutput:", r)
		}
	}()
	cst.cs.applyFileContracts(pb, txn0) // File contract was consumed by the first proof.
	cst.cs.applyStorageProofs(pb, txn1)
}

// TestApplyTurtleDexfundInputs probes the applyTurtleDexfundInputs method of the consensus
// set.
func TestApplyTurtleDexfundInputs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)
	pb.Height = cst.cs.height()

	// Fetch the output id's of each ttdc output in the consensus set.
	var ids []types.TurtleDexfundOutputID
	cst.cs.db.forEachTurtleDexfundOutputs(func(sfoid types.TurtleDexfundOutputID, sfo types.TurtleDexfundOutput) {
		ids = append(ids, sfoid)
	})

	// Apply a transaction with a single siafund input.
	txn := types.Transaction{
		TurtleDexfundInputs: []types.TurtleDexfundInput{
			{ParentID: ids[0]},
		},
	}
	cst.cs.applyTurtleDexfundInputs(pb, txn)
	exists := cst.cs.db.inTurtleDexfundOutputs(ids[0])
	if exists {
		t.Error("Failed to conusme a siafund output")
	}
	if cst.cs.db.lenTurtleDexfundOutputs() != 2 {
		t.Error("siafund outputs not correctly updated", cst.cs.db.lenTurtleDexfundOutputs())
	}
	if len(pb.TurtleDexfundOutputDiffs) != 1 {
		t.Error("block node was not updated for single transaction")
	}
	if pb.TurtleDexfundOutputDiffs[0].Direction != modules.DiffRevert {
		t.Error("wrong diff direction applied when consuming a siafund output")
	}
	if pb.TurtleDexfundOutputDiffs[0].ID != ids[0] {
		t.Error("wrong id used when consuming a siafund output")
	}
	if cst.cs.db.lenDelayedTurtleDexcoinOutputsHeight(cst.cs.height()+types.MaturityDelay) != 2 { // 1 for a block subsidy, 1 for the siafund claim.
		t.Error("siafund claim was not created")
	}
}

// TestMisuseApplyTurtleDexfundInputs misuses applyTurtleDexfundInputs and checks that a
// panic was triggered.
func TestMisuseApplyTurtleDexfundInputs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)
	pb.Height = cst.cs.height()

	// Fetch the output id's of each ttdc output in the consensus set.
	var ids []types.TurtleDexfundOutputID
	cst.cs.db.forEachTurtleDexfundOutputs(func(sfoid types.TurtleDexfundOutputID, sfo types.TurtleDexfundOutput) {
		ids = append(ids, sfoid)
	})

	// Apply a transaction with a single siafund input.
	txn := types.Transaction{
		TurtleDexfundInputs: []types.TurtleDexfundInput{
			{ParentID: ids[0]},
		},
	}
	cst.cs.applyTurtleDexfundInputs(pb, txn)

	// Trigger the panic that occurs when an output is applied incorrectly, and
	// perform a catch to read the error that is created.
	defer func() {
		r := recover()
		if r != ErrMisuseApplyTurtleDexfundInput {
			t.Error("no panic occurred when misusing applyTurtleDexcoinInput")
			t.Error(r)
		}
	}()
	cst.cs.applyTurtleDexfundInputs(pb, txn)
}

// TestApplyTurtleDexfundOutputs probes the applyTurtleDexfundOutputs method of the
// consensus set.
func TestApplyTurtleDexfundOutputs(t *testing.T) {
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
	cst.cs.siafundPool = types.NewCurrency64(101)

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with a single siafund output.
	txn := types.Transaction{
		TurtleDexfundOutputs: []types.TurtleDexfundOutput{{}},
	}
	cst.cs.applyTurtleDexfundOutputs(pb, txn)
	sfoid := txn.TurtleDexfundOutputID(0)
	exists := cst.cs.db.inTurtleDexfundOutputs(sfoid)
	if !exists {
		t.Error("Failed to create siafund output")
	}
	if cst.cs.db.lenTurtleDexfundOutputs() != 4 {
		t.Error("siafund outputs not correctly updated")
	}
	if len(pb.TurtleDexfundOutputDiffs) != 1 {
		t.Error("block node was not updated for single element transaction")
	}
	if pb.TurtleDexfundOutputDiffs[0].Direction != modules.DiffApply {
		t.Error("wrong diff direction applied when creating a siafund output")
	}
	if pb.TurtleDexfundOutputDiffs[0].ID != sfoid {
		t.Error("wrong id used when creating a siafund output")
	}
	if pb.TurtleDexfundOutputDiffs[0].TurtleDexfundOutput.ClaimStart.Cmp64(101) != 0 {
		t.Error("claim start set incorrectly when creating a siafund output")
	}

	// Apply a transaction with 2 ttdc outputs.
	txn = types.Transaction{
		TurtleDexfundOutputs: []types.TurtleDexfundOutput{
			{Value: types.NewCurrency64(1)},
			{Value: types.NewCurrency64(2)},
		},
	}
	cst.cs.applyTurtleDexfundOutputs(pb, txn)
	sfoid0 := txn.TurtleDexfundOutputID(0)
	sfoid1 := txn.TurtleDexfundOutputID(1)
	exists = cst.cs.db.inTurtleDexfundOutputs(sfoid0)
	if !exists {
		t.Error("Failed to create siafund output")
	}
	exists = cst.cs.db.inTurtleDexfundOutputs(sfoid1)
	if !exists {
		t.Error("Failed to create siafund output")
	}
	if cst.cs.db.lenTurtleDexfundOutputs() != 6 {
		t.Error("siafund outputs not correctly updated")
	}
	if len(pb.TurtleDexfundOutputDiffs) != 3 {
		t.Error("block node was not updated for single element transaction")
	}
}

// TestMisuseApplyTurtleDexfundOutputs misuses applyTurtleDexfundOutputs and checks that a
// panic was triggered.
func TestMisuseApplyTurtleDexfundOutputs(t *testing.T) {
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

	// Create a block node to use with application.
	pb := new(processedBlock)

	// Apply a transaction with a single ttdc output.
	txn := types.Transaction{
		TurtleDexfundOutputs: []types.TurtleDexfundOutput{{}},
	}
	cst.cs.applyTurtleDexfundOutputs(pb, txn)

	// Trigger the panic that occurs when an output is applied incorrectly, and
	// perform a catch to read the error that is created.
	defer func() {
		r := recover()
		if r != ErrMisuseApplyTurtleDexfundOutput {
			t.Error("no panic occurred when misusing applyTurtleDexfundInput")
		}
	}()
	cst.cs.applyTurtleDexfundOutputs(pb, txn)
}
*/

// TestApplyArbitraryData probes the applyArbitraryData function.
func TestApplyArbitraryData(t *testing.T) {
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

	apply := func(txn types.Transaction, height types.BlockHeight) {
		err := cst.cs.db.Update(func(tx *bolt.Tx) error {
			// applyArbitraryData expects a BlockPath entry at this height
			tx.Bucket(BlockPath).Put(encoding.Marshal(height), encoding.Marshal(types.BlockID{}))
			applyArbitraryData(tx, &processedBlock{Height: height}, txn)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	addrsChanged := func() bool {
		p, f := cst.cs.FoundationUnlockHashes()
		return p != types.InitialFoundationUnlockHash || f != types.InitialFoundationFailsafeUnlockHash
	}

	// Apply an empty transaction
	apply(types.Transaction{}, types.FoundationHardforkHeight)
	if addrsChanged() {
		t.Error("addrs should not have changed after applying empty txn")
	}

	// Apply data with an invalid prefix -- it should be ignored
	data := encoding.MarshalAll(types.Specifier{'f', 'o', 'o'}, types.FoundationUnlockHashUpdate{})
	apply(types.Transaction{ArbitraryData: [][]byte{data}}, types.FoundationHardforkHeight)
	if addrsChanged() {
		t.Error("addrs should not have changed after applying invalid txn")
	}

	// Apply a validate update before the hardfork -- it should be ignored
	update := types.FoundationUnlockHashUpdate{
		NewPrimary:  types.UnlockHash{1, 2, 3},
		NewFailsafe: types.UnlockHash{4, 5, 6},
	}
	data = encoding.MarshalAll(types.SpecifierFoundation, update)
	apply(types.Transaction{ArbitraryData: [][]byte{data}}, types.FoundationHardforkHeight-1)
	if addrsChanged() {
		t.Fatal("applying valid update before hardfork should not change unlock hashes")
	}
	// Apply the update after the hardfork
	apply(types.Transaction{ArbitraryData: [][]byte{data}}, types.FoundationHardforkHeight)
	if !addrsChanged() {
		t.Fatal("applying valid update did not change unlock hashes")
	}
	// Check that database was updated correctly
	if newPrimary, newFailsafe := cst.cs.FoundationUnlockHashes(); newPrimary != update.NewPrimary || newFailsafe != update.NewFailsafe {
		t.Error("applying valid update did not change unlock hashes")
	}

	// Apply a transaction with two updates; only the first should be applied
	up1 := types.FoundationUnlockHashUpdate{
		NewPrimary:  types.UnlockHash{1, 1, 1},
		NewFailsafe: types.UnlockHash{2, 2, 2},
	}
	up2 := types.FoundationUnlockHashUpdate{
		NewPrimary:  types.UnlockHash{3, 3, 3},
		NewFailsafe: types.UnlockHash{4, 4, 4},
	}
	data1 := encoding.MarshalAll(types.SpecifierFoundation, up1)
	data2 := encoding.MarshalAll(types.SpecifierFoundation, up2)
	apply(types.Transaction{ArbitraryData: [][]byte{data1, data2}}, types.FoundationHardforkHeight+1)
	if newPrimary, newFailsafe := cst.cs.FoundationUnlockHashes(); newPrimary != up1.NewPrimary || newFailsafe != up1.NewFailsafe {
		t.Error("applying two updates did not apply only the first", newPrimary, newFailsafe)
	}
}
