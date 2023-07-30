package mempool

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MysGate/go-fundamental/chain/btc/btcapi"
	"github.com/MysGate/go-fundamental/util"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type UTXO struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Value int64 `json:"value"`
}

// UTXOs is a slice of UTXO
type UTXOs []UTXO

func (c *MempoolClient) ListUnspent(address btcutil.Address) ([]*btcapi.UnspentOutput, error) {
	res, err := c.request(http.MethodGet, fmt.Sprintf("/address/%s/utxo", address.EncodeAddress()), nil)
	if err != nil {
		return nil, err
	}

	var utxos UTXOs
	err = json.Unmarshal(res, &utxos)
	if err != nil {
		return nil, err
	}

	unspentOutputs := make([]*btcapi.UnspentOutput, 0)
	for _, utxo := range utxos {
		txHash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return nil, err
		}
		unspentOutputs = append(unspentOutputs, &btcapi.UnspentOutput{
			Outpoint: wire.NewOutPoint(txHash, uint32(utxo.Vout)),
			Output:   wire.NewTxOut(utxo.Value, address.ScriptAddress()),
		})
	}
	return unspentOutputs, nil
}

func (c *MempoolClient) GetBalance(address btcutil.Address) (int64, error) {
	outputs, err := c.ListUnspent(address)
	if err != nil {
		errMsg := fmt.Sprintf("GetBalance err:%+v", err)
		util.Logger().Error(errMsg)
		return 0, err
	}

	var value int64
	for _, o := range outputs {
		value += o.Output.Value
	}

	return value, nil
}
