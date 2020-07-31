package fees

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

// GetDetailsFunc is a function which looks up transactions by hash.
type GetDetailsFunc func(hash *chainhash.Hash) (*btcjson.TxRawResult, error)

// CalculateFee returns the total fees for the transaction provided.
// TODO(carla): identify change address and split fees between outputs.
func CalculateFee(details GetDetailsFunc, txid *chainhash.Hash) (btcutil.Amount,
	error) {

	var fees btcutil.Amount

	tx, err := details(txid)
	if err != nil {
		return 0, err
	}

	// First, we minus total outputs from our fees.
	for _, out := range tx.Vout {
		amt, err := btcutil.NewAmount(out.Value)
		if err != nil {
			return 0, err
		}

		fees -= amt
	}

	// Next, we lookup each of our inputs to figure out their values and
	// minus them from our fees
	for _, in := range tx.Vin {
		prevOutHash, err := chainhash.NewHashFromStr(in.Txid)
		if err != nil {
			return 0, err
		}

		tx, err := details(prevOutHash)
		if err != nil {
			return 0, err
		}

		prevOut := tx.Vout[in.Vout]
		amt, err := btcutil.NewAmount(prevOut.Value)
		if err != nil {
			return 0, err
		}

		fees += amt
	}

	// Our fees are simply the difference between our input and output
	// total.
	return fees, nil
}
