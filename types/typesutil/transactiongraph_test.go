package typesutil

import (
	"testing"

	"github.com/turtledex/TurtleDexCore/types"

	"github.com/turtledex/errors"
)

// TestTransactionGraph will check that the basic construction of a transaction
// graph works as expected.
func TestTransactionGraph(t *testing.T) {
	// Make a basic transaction.
	var source types.TurtleDexcoinOutputID
	tg := NewTransactionGraph()
	index, err := tg.AddTurtleDexcoinSource(source, types.TurtleDexcoinPrecision.Mul64(3))
	if err != nil {
		t.Fatal(err)
	}
	_, err = tg.AddTurtleDexcoinSource(source, types.TurtleDexcoinPrecision.Mul64(3))
	if !errors.Contains(err, ErrTurtleDexcoinSourceAlreadyAdded) {
		t.Fatal("should not be able to add the same ttdc input source multiple times")
	}
	newIndexes, err := tg.AddTransaction(SimpleTransaction{
		TurtleDexcoinInputs:  []int{index},
		TurtleDexcoinOutputs: []types.Currency{types.TurtleDexcoinPrecision.Mul64(2)},
		MinerFees:      []types.Currency{types.TurtleDexcoinPrecision},
	})
	if err != nil {
		t.Fatal(err)
	}
	txns := tg.Transactions()
	if len(txns) != 1 {
		t.Fatal("expected to get one transaction")
	}
	// Check that the transaction is standalone valid.
	err = txns[0].StandaloneValid(0)
	if err != nil {
		t.Fatal("transactions produced by graph should be valid")
	}

	// Try to build a transaction that has a value mismatch, ensure there is an
	// error.
	_, err = tg.AddTransaction(SimpleTransaction{
		TurtleDexcoinInputs:  []int{newIndexes[0]},
		TurtleDexcoinOutputs: []types.Currency{types.TurtleDexcoinPrecision.Mul64(2)},
		MinerFees:      []types.Currency{types.TurtleDexcoinPrecision},
	})
	if !errors.Contains(err, ErrTurtleDexcoinInputsOutputsMismatch) {
		t.Fatal("An error should be returned when a transaction's outputs and inputs mismatch")
	}
	_, err = tg.AddTransaction(SimpleTransaction{
		TurtleDexcoinInputs:  []int{2},
		TurtleDexcoinOutputs: []types.Currency{types.TurtleDexcoinPrecision},
		MinerFees:      []types.Currency{types.TurtleDexcoinPrecision},
	})
	if !errors.Contains(err, ErrNoSuchTurtleDexcoinInput) {
		t.Fatal("An error should be returned when a transaction spends a missing input")
	}
	_, err = tg.AddTransaction(SimpleTransaction{
		TurtleDexcoinInputs:  []int{0},
		TurtleDexcoinOutputs: []types.Currency{types.TurtleDexcoinPrecision},
		MinerFees:      []types.Currency{types.TurtleDexcoinPrecision},
	})
	if !errors.Contains(err, ErrTurtleDexcoinInputAlreadyUsed) {
		t.Fatal("Error should be returned when a transaction spends an input that has been spent before")
	}

	// Build a correct second transaction, see that it validates.
	_, err = tg.AddTransaction(SimpleTransaction{
		TurtleDexcoinInputs:  []int{newIndexes[0]},
		TurtleDexcoinOutputs: []types.Currency{types.TurtleDexcoinPrecision},
		MinerFees:      []types.Currency{types.TurtleDexcoinPrecision},
	})
	if err != nil {
		t.Fatal("Transaction was built incorrectly", err)
	}
}
