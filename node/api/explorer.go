package api

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/turtledex/TurtleDexCore/build"
	"github.com/turtledex/TurtleDexCore/crypto"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/types"
)

type (
	// ExplorerBlock is a block with some extra information such as the id and
	// height. This information is provided for programs that may not be
	// complex enough to compute the ID on their own.
	ExplorerBlock struct {
		MinerPayoutIDs []types.TurtleDexcoinOutputID `json:"minerpayoutids"`
		Transactions   []ExplorerTransaction   `json:"transactions"`
		RawBlock       types.Block             `json:"rawblock"`

		modules.BlockFacts
	}

	// ExplorerTransaction is a transcation with some extra information such as
	// the parent block. This information is provided for programs that may not
	// be complex enough to compute the extra information on their own.
	ExplorerTransaction struct {
		ID             types.TransactionID `json:"id"`
		Height         types.BlockHeight   `json:"height"`
		Parent         types.BlockID       `json:"parent"`
		RawTransaction types.Transaction   `json:"rawtransaction"`

		TurtleDexcoinInputOutputs                      []types.TurtleDexcoinOutput     `json:"ttdcinputoutputs"` // the outputs being spent
		TurtleDexcoinOutputIDs                         []types.TurtleDexcoinOutputID   `json:"ttdcoutputids"`
		FileContractIDs                          []types.FileContractID    `json:"filecontractids"`
		FileContractValidProofOutputIDs          [][]types.TurtleDexcoinOutputID `json:"filecontractvalidproofoutputids"`          // outer array is per-contract
		FileContractMissedProofOutputIDs         [][]types.TurtleDexcoinOutputID `json:"filecontractmissedproofoutputids"`         // outer array is per-contract
		FileContractRevisionValidProofOutputIDs  [][]types.TurtleDexcoinOutputID `json:"filecontractrevisionvalidproofoutputids"`  // outer array is per-revision
		FileContractRevisionMissedProofOutputIDs [][]types.TurtleDexcoinOutputID `json:"filecontractrevisionmissedproofoutputids"` // outer array is per-revision
		StorageProofOutputIDs                    [][]types.TurtleDexcoinOutputID `json:"storageproofoutputids"`                    // outer array is per-payout
		StorageProofOutputs                      [][]types.TurtleDexcoinOutput   `json:"storageproofoutputs"`                      // outer array is per-payout
		TurtleDexfundInputOutputs                      []types.TurtleDexfundOutput     `json:"siafundinputoutputs"`                      // the outputs being spent
		TurtleDexfundOutputIDs                         []types.TurtleDexfundOutputID   `json:"siafundoutputids"`
		TurtleDexfundClaimOutputIDs                    []types.TurtleDexcoinOutputID   `json:"siafundclaimoutputids"`
	}

	// ExplorerGET is the object returned as a response to a GET request to
	// /explorer.
	ExplorerGET struct {
		modules.BlockFacts
	}

	// ExplorerBlockGET is the object returned by a GET request to
	// /explorer/block.
	ExplorerBlockGET struct {
		Block ExplorerBlock `json:"block"`
	}

	// ExplorerHashGET is the object returned as a response to a GET request to
	// /explorer/hash. The HashType will indicate whether the hash corresponds
	// to a block id, a transaction id, a ttdc output id, a file contract
	// id, or a siafund output id. In the case of a block id, 'Block' will be
	// filled out and all the rest of the fields will be blank. In the case of
	// a transaction id, 'Transaction' will be filled out and all the rest of
	// the fields will be blank. For everything else, 'Transactions' and
	// 'Blocks' will/may be filled out and everything else will be blank.
	ExplorerHashGET struct {
		HashType     string                `json:"hashtype"`
		Block        ExplorerBlock         `json:"block"`
		Blocks       []ExplorerBlock       `json:"blocks"`
		Transaction  ExplorerTransaction   `json:"transaction"`
		Transactions []ExplorerTransaction `json:"transactions"`
	}
)

// buildExplorerTransaction takes a transaction and the height + id of the
// block it appears in an uses that to build an explorer transaction.
func (api *API) buildExplorerTransaction(height types.BlockHeight, parent types.BlockID, txn types.Transaction) (et ExplorerTransaction) {
	// Get the header information for the transaction.
	et.ID = txn.ID()
	et.Height = height
	et.Parent = parent
	et.RawTransaction = txn

	// Add the ttdc outputs that correspond with each ttdc input.
	for _, sci := range txn.TurtleDexcoinInputs {
		sco, exists := api.explorer.TurtleDexcoinOutput(sci.ParentID)
		if build.DEBUG && !exists {
			panic("could not find corresponding ttdc output")
		}
		et.TurtleDexcoinInputOutputs = append(et.TurtleDexcoinInputOutputs, sco)
	}

	for i := range txn.TurtleDexcoinOutputs {
		et.TurtleDexcoinOutputIDs = append(et.TurtleDexcoinOutputIDs, txn.TurtleDexcoinOutputID(uint64(i)))
	}

	// Add all of the valid and missed proof ids as extra data to the file
	// contracts.
	for i, fc := range txn.FileContracts {
		fcid := txn.FileContractID(uint64(i))
		var fcvpoids []types.TurtleDexcoinOutputID
		var fcmpoids []types.TurtleDexcoinOutputID
		for j := range fc.ValidProofOutputs {
			fcvpoids = append(fcvpoids, fcid.StorageProofOutputID(types.ProofValid, uint64(j)))
		}
		for j := range fc.MissedProofOutputs {
			fcmpoids = append(fcmpoids, fcid.StorageProofOutputID(types.ProofMissed, uint64(j)))
		}
		et.FileContractIDs = append(et.FileContractIDs, fcid)
		et.FileContractValidProofOutputIDs = append(et.FileContractValidProofOutputIDs, fcvpoids)
		et.FileContractMissedProofOutputIDs = append(et.FileContractMissedProofOutputIDs, fcmpoids)
	}

	// Add all of the valid and missed proof ids as extra data to the file
	// contract revisions.
	for _, fcr := range txn.FileContractRevisions {
		var fcrvpoids []types.TurtleDexcoinOutputID
		var fcrmpoids []types.TurtleDexcoinOutputID
		for j := range fcr.NewValidProofOutputs {
			fcrvpoids = append(fcrvpoids, fcr.ParentID.StorageProofOutputID(types.ProofValid, uint64(j)))
		}
		for j := range fcr.NewMissedProofOutputs {
			fcrmpoids = append(fcrmpoids, fcr.ParentID.StorageProofOutputID(types.ProofMissed, uint64(j)))
		}
		et.FileContractValidProofOutputIDs = append(et.FileContractValidProofOutputIDs, fcrvpoids)
		et.FileContractMissedProofOutputIDs = append(et.FileContractMissedProofOutputIDs, fcrmpoids)
	}

	// Add all of the output ids and outputs corresponding with each storage
	// proof.
	for _, sp := range txn.StorageProofs {
		fileContract, fileContractRevisions, fileContractExists, _ := api.explorer.FileContractHistory(sp.ParentID)
		if !fileContractExists && build.DEBUG {
			panic("could not find a file contract connected with a storage proof")
		}
		var storageProofOutputs []types.TurtleDexcoinOutput
		if len(fileContractRevisions) > 0 {
			storageProofOutputs = fileContractRevisions[len(fileContractRevisions)-1].NewValidProofOutputs
		} else {
			storageProofOutputs = fileContract.ValidProofOutputs
		}
		var storageProofOutputIDs []types.TurtleDexcoinOutputID
		for i := range storageProofOutputs {
			storageProofOutputIDs = append(storageProofOutputIDs, sp.ParentID.StorageProofOutputID(types.ProofValid, uint64(i)))
		}
		et.StorageProofOutputIDs = append(et.StorageProofOutputIDs, storageProofOutputIDs)
		et.StorageProofOutputs = append(et.StorageProofOutputs, storageProofOutputs)
	}

	// Add the siafund outputs that correspond to each ttdc input.
	for _, sci := range txn.TurtleDexfundInputs {
		sco, exists := api.explorer.TurtleDexfundOutput(sci.ParentID)
		if build.DEBUG && !exists {
			panic("could not find corresponding siafund output")
		}
		et.TurtleDexfundInputOutputs = append(et.TurtleDexfundInputOutputs, sco)
	}

	for i := range txn.TurtleDexfundOutputs {
		et.TurtleDexfundOutputIDs = append(et.TurtleDexfundOutputIDs, txn.TurtleDexfundOutputID(uint64(i)))
	}

	for _, sfi := range txn.TurtleDexfundInputs {
		et.TurtleDexfundClaimOutputIDs = append(et.TurtleDexfundClaimOutputIDs, sfi.ParentID.TurtleDexClaimOutputID())
	}
	return et
}

// buildExplorerBlock takes a block and its height and uses it to construct an
// explorer block.
func (api *API) buildExplorerBlock(height types.BlockHeight, block types.Block) ExplorerBlock {
	var mpoids []types.TurtleDexcoinOutputID
	for i := range block.MinerPayouts {
		mpoids = append(mpoids, block.MinerPayoutID(uint64(i)))
	}

	var etxns []ExplorerTransaction
	for _, txn := range block.Transactions {
		etxns = append(etxns, api.buildExplorerTransaction(height, block.ID(), txn))
	}

	facts, exists := api.explorer.BlockFacts(height)
	if build.DEBUG && !exists {
		panic("incorrect request to buildExplorerBlock - block does not exist")
	}

	return ExplorerBlock{
		MinerPayoutIDs: mpoids,
		Transactions:   etxns,
		RawBlock:       block,

		BlockFacts: facts,
	}
}

// explorerHandler handles API calls to /explorer/blocks/:height.
func (api *API) explorerBlocksHandler(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	// Parse the height that's being requested.
	var height types.BlockHeight
	_, err := fmt.Sscan(ps.ByName("height"), &height)
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	// Fetch and return the explorer block.
	block, exists := api.cs.BlockAtHeight(height)
	if !exists {
		WriteError(w, Error{"no block found at input height in call to /explorer/block"}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, ExplorerBlockGET{
		Block: api.buildExplorerBlock(height, block),
	})
}

// buildTransactionSet returns the blocks and transactions that are associated
// with a set of transaction ids.
func (api *API) buildTransactionSet(txids []types.TransactionID) (txns []ExplorerTransaction, blocks []ExplorerBlock) {
	for _, txid := range txids {
		// Get the block containing the transaction - in the case of miner
		// payouts, the block might be the transaction.
		block, height, exists := api.explorer.Transaction(txid)
		if !exists && build.DEBUG {
			panic("explorer pointing to nonexistent txn")
		}

		// Check if the block is the transaction.
		if types.TransactionID(block.ID()) == txid {
			blocks = append(blocks, api.buildExplorerBlock(height, block))
		} else {
			// Find the transaction within the block with the correct id.
			for _, t := range block.Transactions {
				if t.ID() == txid {
					txns = append(txns, api.buildExplorerTransaction(height, block.ID(), t))
					break
				}
			}
		}
	}
	return txns, blocks
}

// explorerHashHandler handles GET requests to /explorer/hash/:hash.
func (api *API) explorerHashHandler(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	// Scan the hash as a hash. If that fails, try scanning the hash as an
	// address.
	hash, err := scanHash(ps.ByName("hash"))
	if err != nil {
		addr, err := scanAddress(ps.ByName("hash"))
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		hash = crypto.Hash(addr)
	}

	// TODO: lookups on the zero hash are too expensive to allow. Need a
	// better way to handle this case.
	if hash == (crypto.Hash{}) {
		WriteError(w, Error{"can't lookup the empty unlock hash"}, http.StatusBadRequest)
		return
	}

	// Try the hash as a block id.
	block, height, exists := api.explorer.Block(types.BlockID(hash))
	if exists {
		WriteJSON(w, ExplorerHashGET{
			HashType: "blockid",
			Block:    api.buildExplorerBlock(height, block),
		})
		return
	}

	// Try the hash as a transaction id.
	block, height, exists = api.explorer.Transaction(types.TransactionID(hash))
	if exists {
		var txn types.Transaction
		for _, t := range block.Transactions {
			if t.ID() == types.TransactionID(hash) {
				txn = t
			}
		}
		WriteJSON(w, ExplorerHashGET{
			HashType:    "transactionid",
			Transaction: api.buildExplorerTransaction(height, block.ID(), txn),
		})
		return
	}

	// Try the hash as a ttdc output id.
	txids := api.explorer.TurtleDexcoinOutputID(types.TurtleDexcoinOutputID(hash))
	if len(txids) != 0 {
		txns, blocks := api.buildTransactionSet(txids)
		WriteJSON(w, ExplorerHashGET{
			HashType:     "ttdcoutputid",
			Blocks:       blocks,
			Transactions: txns,
		})
		return
	}

	// Try the hash as a file contract id.
	txids = api.explorer.FileContractID(types.FileContractID(hash))
	if len(txids) != 0 {
		txns, blocks := api.buildTransactionSet(txids)
		WriteJSON(w, ExplorerHashGET{
			HashType:     "filecontractid",
			Blocks:       blocks,
			Transactions: txns,
		})
		return
	}

	// Try the hash as a siafund output id.
	txids = api.explorer.TurtleDexfundOutputID(types.TurtleDexfundOutputID(hash))
	if len(txids) != 0 {
		txns, blocks := api.buildTransactionSet(txids)
		WriteJSON(w, ExplorerHashGET{
			HashType:     "siafundoutputid",
			Blocks:       blocks,
			Transactions: txns,
		})
		return
	}

	// Try the hash as an unlock hash. Unlock hash is checked last because
	// unlock hashes do not have collision-free guarantees. Someone can create
	// an unlock hash that collides with another object id. They will not be
	// able to use the unlock hash, but they can disrupt the explorer. This is
	// handled by checking the unlock hash last. Anyone intentionally creating
	// a colliding unlock hash (such a collision can only happen if done
	// intentionally) will be unable to find their unlock hash in the
	// blockchain through the explorer hash lookup.
	txids = api.explorer.UnlockHash(types.UnlockHash(hash))
	if len(txids) != 0 {
		txns, blocks := api.buildTransactionSet(txids)
		WriteJSON(w, ExplorerHashGET{
			HashType:     "unlockhash",
			Blocks:       blocks,
			Transactions: txns,
		})
		return
	}

	// Hash not found, return an error.
	WriteError(w, Error{"unrecognized hash used as input to /explorer/hash"}, http.StatusBadRequest)
}

// explorerHandler handles API calls to /explorer
func (api *API) explorerHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	facts := api.explorer.LatestBlockFacts()
	WriteJSON(w, ExplorerGET{
		BlockFacts: facts,
	})
}
