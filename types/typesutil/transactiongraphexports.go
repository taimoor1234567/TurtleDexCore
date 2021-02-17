package typesutil

import (
	"fmt"

	"github.com/turtledex/TurtleDexCore/types"

	"github.com/turtledex/errors"
)

var (
	// AnyoneCanSpendUnlockHash is the unlock hash of unlock conditions that are
	// trivially spendable.
	AnyoneCanSpendUnlockHash types.UnlockHash = types.UnlockConditions{}.UnlockHash()
)

var (
	// ErrTurtleDexcoinSourceAlreadyAdded is the error returned when a user tries to
	// provide the same source ttdc input multiple times.
	ErrTurtleDexcoinSourceAlreadyAdded = errors.New("source ttdc input has already been used")

	// ErrTurtleDexcoinInputAlreadyUsed warns a user that a ttdc input has already
	// been used in the transaction graph.
	ErrTurtleDexcoinInputAlreadyUsed = errors.New("cannot use the same ttdc input twice in a graph")

	// ErrNoSuchTurtleDexcoinInput warns a user that they are trying to reference a
	// ttdc input which does not yet exist.
	ErrNoSuchTurtleDexcoinInput = errors.New("no ttdc input exists with that index")

	// ErrTurtleDexcoinInputsOutputsMismatch warns a user that they have constructed a
	// transaction which does not spend the same amount of ttdcs that it
	// consumes.
	ErrTurtleDexcoinInputsOutputsMismatch = errors.New("ttdc input value to transaction does not match ttdc output value of transaction")
)

// ttdcInput defines a ttdc input within the transaction graph, containing
// the input itself, the value of the input, and a flag indicating whether or
// not the input has been used within the graph already.
type ttdcInput struct {
	input types.TurtleDexcoinInput
	used  bool
	value types.Currency
}

// TransactionGraph is a helper tool to allow a user to easily construct
// elaborate transaction graphs. The transaction tool will handle creating valid
// transactions, providing the user with a clean interface for building
// transactions.
type TransactionGraph struct {
	// A map that tracks which source inputs have been consumed, to double check
	// that the user is not supplying the same source inputs multiple times.
	usedTurtleDexcoinInputSources map[types.TurtleDexcoinOutputID]struct{}

	ttdcInputs []ttdcInput

	transactions []types.Transaction
}

// SimpleTransaction specifies what outputs it spends, and what outputs it
// creates, by index. When passed in TransactionGraph, it will be automatically
// transformed into a valid transaction.
//
// Currently, there is only support for TurtleDexcoinInputs, TurtleDexcoinOutputs, and
// MinerFees, however the code has been structured so that support for TurtleDexfunds
// and FileContracts can be easily added in the future.
type SimpleTransaction struct {
	TurtleDexcoinInputs  []int            // Which inputs to use, by index.
	TurtleDexcoinOutputs []types.Currency // The values of each output.

	/*
		TurtleDexfundInputs  []int            // Which inputs to use, by index.
		TurtleDexfundOutputs []types.Currency // The values of each output.

		FileContracts         int   // The number of file contracts to create.
		FileContractRevisions []int // Which file contracts to revise.
		StorageProofs         []int // Which file contracts to create proofs for.
	*/

	MinerFees []types.Currency // The fees used.

	/*
		ArbitraryData [][]byte // Arbitrary data to include in the transaction.
	*/
}

// AddTurtleDexcoinSource will add a new source of ttdcs to the transaction graph,
// returning the index that this source can be referenced by. The provided
// output must have the address AnyoneCanSpendUnlockHash.
//
// The value is used as an input so that the graph can check whether all
// transactions are spending as many ttdcs as they create.
func (tg *TransactionGraph) AddTurtleDexcoinSource(scoid types.TurtleDexcoinOutputID, value types.Currency) (int, error) {
	// Check if this scoid has already been used.
	_, exists := tg.usedTurtleDexcoinInputSources[scoid]
	if exists {
		return -1, ErrTurtleDexcoinSourceAlreadyAdded
	}

	i := len(tg.ttdcInputs)
	tg.ttdcInputs = append(tg.ttdcInputs, ttdcInput{
		input: types.TurtleDexcoinInput{
			ParentID: scoid,
		},
		value: value,
	})
	tg.usedTurtleDexcoinInputSources[scoid] = struct{}{}
	return i, nil
}

// AddTransaction will add a new transaction to the transaction graph, following
// the guide of the input. The indexes of all the outputs created will be
// returned.
func (tg *TransactionGraph) AddTransaction(st SimpleTransaction) (newTurtleDexcoinInputs []int, err error) {
	var txn types.Transaction
	var totalIn types.Currency
	var totalOut types.Currency

	// Consume all of the inputs.
	for _, sci := range st.TurtleDexcoinInputs {
		if sci >= len(tg.ttdcInputs) {
			return nil, ErrNoSuchTurtleDexcoinInput
		}
		if tg.ttdcInputs[sci].used {
			return nil, ErrTurtleDexcoinInputAlreadyUsed
		}
		txn.TurtleDexcoinInputs = append(txn.TurtleDexcoinInputs, tg.ttdcInputs[sci].input)
		totalIn = totalIn.Add(tg.ttdcInputs[sci].value)
	}

	// Create all of the outputs.
	for _, scov := range st.TurtleDexcoinOutputs {
		txn.TurtleDexcoinOutputs = append(txn.TurtleDexcoinOutputs, types.TurtleDexcoinOutput{
			UnlockHash: AnyoneCanSpendUnlockHash,
			Value:      scov,
		})
		totalOut = totalOut.Add(scov)
	}

	// Add all of the fees.
	txn.MinerFees = st.MinerFees
	for _, fee := range st.MinerFees {
		totalOut = totalOut.Add(fee)
	}

	// Check that the transaction is consistent.
	if totalIn.Cmp(totalOut) != 0 {
		valuesErr := fmt.Errorf("total input: %s, total output: %s", totalIn, totalOut)
		extendedErr := errors.Extend(ErrTurtleDexcoinInputsOutputsMismatch, valuesErr)
		return nil, extendedErr
	}

	// Update the set of ttdc inputs that have been used successfully. This
	// must be done after all error checking is complete.
	for _, sci := range st.TurtleDexcoinInputs {
		tg.ttdcInputs[sci].used = true
	}
	tg.transactions = append(tg.transactions, txn)
	for i, sco := range txn.TurtleDexcoinOutputs {
		newTurtleDexcoinInputs = append(newTurtleDexcoinInputs, len(tg.ttdcInputs))
		tg.ttdcInputs = append(tg.ttdcInputs, ttdcInput{
			input: types.TurtleDexcoinInput{
				ParentID: txn.TurtleDexcoinOutputID(uint64(i)),
			},
			value: sco.Value,
		})
	}
	return newTurtleDexcoinInputs, nil
}

// Transactions will return the transactions that were built up in the graph.
func (tg *TransactionGraph) Transactions() []types.Transaction {
	return tg.transactions
}

// NewTransactionGraph will return a blank transaction graph that is ready for
// use.
func NewTransactionGraph() *TransactionGraph {
	return &TransactionGraph{
		usedTurtleDexcoinInputSources: make(map[types.TurtleDexcoinOutputID]struct{}),
	}
}
