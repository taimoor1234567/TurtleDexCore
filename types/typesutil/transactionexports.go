package typesutil

import (
	"fmt"
	"strings"

	"github.com/turtledex/TurtleDexCore/crypto"
	"github.com/turtledex/TurtleDexCore/types"
)

// SprintTxnWithObjectIDs creates a string representing this Transaction in human-readable form with all
// object IDs included to allow for easy dependency matching (by humans) in
// debug-logs.
func SprintTxnWithObjectIDs(t types.Transaction) string {
	var str strings.Builder
	txIDString := crypto.Hash(t.ID()).String()
	fmt.Fprintf(&str, "\nTransaction ID: %s", txIDString)

	if len(t.TurtleDexcoinInputs) != 0 {
		fmt.Fprintf(&str, "\nTurtleDexcoinInputs:\n")
		for i, input := range t.TurtleDexcoinInputs {
			parentIDString := crypto.Hash(input.ParentID).String()
			fmt.Fprintf(&str, "\t%d: %s\n", i, parentIDString)
		}
	}
	if len(t.TurtleDexcoinOutputs) != 0 {
		fmt.Fprintf(&str, "TurtleDexcoinOutputs:\n")
		for i := range t.TurtleDexcoinOutputs {
			oidString := crypto.Hash(t.TurtleDexcoinOutputID(uint64(i))).String()
			fmt.Fprintf(&str, "\t%d: %s\n", i, oidString)
		}
	}
	if len(t.FileContracts) != 0 {
		fmt.Fprintf(&str, "FileContracts:\n")
		for i := range t.FileContracts {
			fcIDString := crypto.Hash(t.FileContractID(uint64(i))).String()
			fmt.Fprintf(&str, "\t%d: %s\n", i, fcIDString)
		}
	}
	if len(t.FileContractRevisions) != 0 {
		fmt.Fprintf(&str, "FileContractRevisions:\n")
		for _, fcr := range t.FileContractRevisions {
			parentIDString := crypto.Hash(fcr.ParentID).String()
			fmt.Fprintf(&str, "\t%d, %s\n", fcr.NewRevisionNumber, parentIDString)
		}
	}
	if len(t.StorageProofs) != 0 {
		fmt.Fprintf(&str, "StorageProofs:\n")
		for _, sp := range t.StorageProofs {
			parentIDString := crypto.Hash(sp.ParentID).String()
			fmt.Fprintf(&str, "\t%s\n", parentIDString)
		}
	}
	if len(t.TurtleDexfundInputs) != 0 {
		fmt.Fprintf(&str, "TurtleDexfundInputs:\n")
		for i, input := range t.TurtleDexfundInputs {
			parentIDString := crypto.Hash(input.ParentID).String()
			fmt.Fprintf(&str, "\t%d: %s\n", i, parentIDString)
		}
	}
	if len(t.TurtleDexfundOutputs) != 0 {
		fmt.Fprintf(&str, "TurtleDexfundOutputs:\n")
		for i := range t.TurtleDexfundOutputs {
			oidString := crypto.Hash(t.TurtleDexfundOutputID(uint64(i))).String()
			fmt.Fprintf(&str, "\t%d: %s\n", i, oidString)
		}
	}
	return str.String()
}
