package siatest

import (
	"math"

	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
)

// ConfirmedBalance returns the confirmed ttdc balance of the node's
// wallet.
func (tn *TestNode) ConfirmedBalance() (types.Currency, error) {
	wg, err := tn.WalletGet()
	return wg.ConfirmedTurtleDexcoinBalance, err
}

// ConfirmedTransactions returns all of the wallet's tracked confirmed
// transactions.
func (tn *TestNode) ConfirmedTransactions() ([]modules.ProcessedTransaction, error) {
	wtg, err := tn.WalletTransactionsGet(0, math.MaxUint64)
	return wtg.ConfirmedTransactions, err
}
