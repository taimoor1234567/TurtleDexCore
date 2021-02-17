package contractor

import (
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
)

// These interfaces define the HostDB's dependencies. Using the smallest
// interface possible makes it easier to mock these dependencies in testing.
type (
	consensusSet interface {
		ConsensusSetSubscribe(modules.ConsensusSetSubscriber, modules.ConsensusChangeID, <-chan struct{}) error
		Synced() bool
		Unsubscribe(modules.ConsensusSetSubscriber)
	}
	// In order to restrict the modules.TransactionBuilder interface, we must
	// provide a shim to bridge the gap between modules.Wallet and
	// transactionBuilder.
	walletShim interface {
		NextAddress() (types.UnlockConditions, error)
		PrimarySeed() (modules.Seed, uint64, error)
		StartTransaction() (modules.TransactionBuilder, error)
		RegisterTransaction(types.Transaction, []types.Transaction) (modules.TransactionBuilder, error)
		Unlocked() (bool, error)
	}
	transactionBuilder interface {
		AddArbitraryData([]byte) uint64
		AddFileContract(types.FileContract) uint64
		AddFileContractRevision(types.FileContractRevision) uint64
		AddMinerFee(types.Currency) uint64
		AddParents([]types.Transaction)
		AddTurtleDexcoinInput(types.TurtleDexcoinInput) uint64
		AddTurtleDexcoinOutput(types.TurtleDexcoinOutput) uint64
		ReplaceTurtleDexcoinOutput(uint64, types.TurtleDexcoinOutput) error
		AddTransactionSignature(types.TransactionSignature) uint64
		Copy() modules.TransactionBuilder
		Drop()
		FundTurtleDexcoins(types.Currency) error
		MarkWalletInputs() bool
		Sign(bool) ([]types.Transaction, error)
		UnconfirmedParents() ([]types.Transaction, error)
		View() (types.Transaction, []types.Transaction)
		ViewAdded() (parents, coins, funds, signatures []int)
	}
	transactionPool interface {
		AcceptTransactionSet([]types.Transaction) error
		FeeEstimation() (min types.Currency, max types.Currency)
	}

	hostDB interface {
		AllHosts() ([]modules.HostDBEntry, error)
		ActiveHosts() ([]modules.HostDBEntry, error)
		CheckForIPViolations([]types.TurtleDexPublicKey) ([]types.TurtleDexPublicKey, error)
		Filter() (modules.FilterMode, map[string]types.TurtleDexPublicKey, error)
		SetFilterMode(fm modules.FilterMode, hosts []types.TurtleDexPublicKey) error
		Host(types.TurtleDexPublicKey) (modules.HostDBEntry, bool, error)
		IncrementSuccessfulInteractions(key types.TurtleDexPublicKey) error
		IncrementFailedInteractions(key types.TurtleDexPublicKey) error
		InitialScanComplete() (complete bool, err error)
		RandomHosts(n int, blacklist, addressBlacklist []types.TurtleDexPublicKey) ([]modules.HostDBEntry, error)
		UpdateContracts([]modules.RenterContract) error
		ScoreBreakdown(modules.HostDBEntry) (modules.HostScoreBreakdown, error)
		SetAllowance(allowance modules.Allowance) error
	}

	persister interface {
		save(contractorPersist) error
		load(*contractorPersist) error
	}
)

// WalletBridge is a bridge for the wallet because wallet is not directly
// compatible with modules.Wallet (wrong type signature for StartTransaction),
// we must provide a bridge type.
type WalletBridge struct {
	W walletShim
}

// NextAddress computes and returns the next address of the wallet.
func (ws *WalletBridge) NextAddress() (types.UnlockConditions, error) { return ws.W.NextAddress() }

// PrimarySeed returns the primary wallet seed.
func (ws *WalletBridge) PrimarySeed() (modules.Seed, uint64, error) { return ws.W.PrimarySeed() }

// StartTransaction creates a new transactionBuilder that can be used to create
// and sign a transaction.
func (ws *WalletBridge) StartTransaction() (transactionBuilder, error) {
	return ws.W.StartTransaction()
}

// RegisterTransaction creates a new transactionBuilder from a transaction and parent transactions.
func (ws *WalletBridge) RegisterTransaction(t types.Transaction, parents []types.Transaction) (transactionBuilder, error) {
	return ws.W.RegisterTransaction(t, parents)
}

// Unlocked reports whether the wallet bridge is unlocked.
func (ws *WalletBridge) Unlocked() (bool, error) { return ws.W.Unlocked() }
