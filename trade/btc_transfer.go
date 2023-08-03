package trade

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/MysGate/go-fundamental/chain/btc/btcapi"
	"github.com/MysGate/go-fundamental/chain/btc/txpool"
	"github.com/MysGate/go-fundamental/util"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type tx_type int

const (
	tx_a_b tx_type = iota
	tx_as_b
	tx_as_bs
)

const (
	dirty_bitcoin int64 = 546
)

type Target struct {
	Addr   btcutil.Address
	Amount int64
}

func BtcWif2Address(wif string, netParams *chaincfg.Params) (address btcutil.Address) {
	privKeyWif, err := btcutil.DecodeWIF(wif)
	if err != nil {
		errMsg := fmt.Sprintf("BtcWif2Address err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	pubKeySerial := privKeyWif.PrivKey.PubKey().SerializeUncompressed()
	pubKey, err := btcutil.NewAddressPubKey(pubKeySerial, netParams)
	if err != nil {
		errMsg := fmt.Sprintf("BtcWif2Address err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	addr := pubKey.EncodeAddress()
	address, err = btcutil.DecodeAddress(addr, netParams)
	if err != nil {
		errMsg := fmt.Sprintf("BtcWif2Address err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}
	return
}

func GetTxFee(tx *wire.MsgTx, l feelevel) int64 {
	feeRate := GetBtcFee().GetFeeRate(l)
	fee := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(tx))) * btcutil.Amount(feeRate)
	return int64(fee)
}

func SignTx(privKey string, pkScripts []string, redeemTx *wire.MsgTx) (*wire.MsgTx, error) {
	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		errMsg := fmt.Sprintf("SignTx err:%+v", err)
		util.Logger().Error(errMsg)
		return nil, err
	}

	for idx, pkScript := range pkScripts {
		sourcePKScript, err := hex.DecodeString(pkScript)
		if err != nil {
			errMsg := fmt.Sprintf("SignTx err:%+v", err)
			util.Logger().Error(errMsg)
			return nil, err
		}

		signature, err := txscript.SignatureScript(redeemTx, idx, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
		if err != nil {
			errMsg := fmt.Sprintf("SignTx err:%+v", err)
			util.Logger().Error(errMsg)
			return nil, err
		}

		redeemTx.TxIn[idx].SignatureScript = signature
	}

	return redeemTx, nil
}

func GetBalance(outputs []*btcapi.UnspentOutput) (amount int64) {
	for _, o := range outputs {
		amount += o.Output.Value
	}
	return
}

func selectA2BOutput(outputs []*btcapi.UnspentOutput, amount int64, last_out *btcapi.UnspentOutput) (output *btcapi.UnspentOutput) {
	var last_idx int = -1
	if last_out != nil {
		last_idx = int(last_out.Outpoint.Index)
	}

	for _, out := range outputs {
		if out.Output.Value > amount && int(out.Outpoint.Index) > last_idx {
			return out
		}
	}

	return nil
}

func selectAS2BSOutputs(outputs []*btcapi.UnspentOutput, amount int64) (outs [][]*btcapi.UnspentOutput) {
	combinations := util.All(outputs)
	for _, combination := range combinations {
		var total int64
		for _, sub := range combination {
			total += sub.Output.Value
		}

		if total > amount {
			outs = append(outs, combination)
		}
	}
	return
}

func buildTxin(redeemTx *wire.MsgTx, out *btcapi.UnspentOutput) (pkScript string) {
	outPoint := wire.NewOutPoint(&out.Outpoint.Hash, out.Outpoint.Index)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)
	pkScript = string(out.Output.PkScript)
	return
}

func buildTxout(redeemTx *wire.MsgTx, target *Target, addTx bool) (*wire.TxOut, error) {
	outScript, err := txscript.PayToAddrScript(target.Addr)
	if err != nil {
		errMsg := fmt.Sprintf("buildMsgTx err:%+v", err)
		err = errors.New(errMsg)
		util.Logger().Error(errMsg)
		return nil, err
	}

	redeemTxOut := wire.NewTxOut(target.Amount, outScript)
	if addTx {
		redeemTx.AddTxOut(redeemTxOut)
	}

	return redeemTxOut, nil
}

func buildMsgTx(t tx_type, receivers []*Target, changeReceiver btcutil.Address, outputs []*btcapi.UnspentOutput) (redeemTx *wire.MsgTx, PKScripts []string, err error) {
	redeemTx = wire.NewMsgTx(wire.TxVersion)
	switch t {
	case tx_a_b:
		var out *btcapi.UnspentOutput
		for {
			out = selectA2BOutput(outputs, receivers[0].Amount, nil)
			if out == nil {
				err = errors.New("buildMsgTx: txin no valid utxo")
				return
			}

			pkScript := buildTxin(redeemTx, out)
			PKScripts = append(PKScripts, pkScript)

			_, err = buildTxout(redeemTx, receivers[0], true)
			if err != nil {
				errMsg := fmt.Sprintf("buildMsgTx err:%+v", err)
				err = errors.New(errMsg)
				util.Logger().Error(errMsg)
				return
			}

			// change
			var redeemTxOutChange *wire.TxOut
			redeemTxOutChange, err = buildTxout(redeemTx, &Target{Addr: changeReceiver}, true)
			if err != nil {
				errMsg := fmt.Sprintf("buildMsgTx err:%+v", err)
				err = errors.New(errMsg)
				util.Logger().Error(errMsg)
				return
			}

			// adjuest fee
			fee := GetTxFee(redeemTx, feelevel_mid)
			if out.Output.Value-fee-receivers[0].Amount < dirty_bitcoin {
				redeemTx.TxIn = nil
				redeemTx.TxOut = nil
				continue
			} else {
				redeemTxOutChange.Value = out.Output.Value - fee - receivers[0].Amount
			}
		}
	case tx_as_b:
		// txin
		for _, output := range outputs {
			pkScript := buildTxin(redeemTx, output)
			PKScripts = append(PKScripts, pkScript)
		}
		// txout
		total := GetBalance(outputs)
		receivers[0].Amount = total
		var redeemTxOut *wire.TxOut
		redeemTxOut, err = buildTxout(redeemTx, receivers[0], true)
		if err != nil {
			errMsg := fmt.Sprintf("buildMsgTx err:%+v", err)
			err = errors.New(errMsg)
			util.Logger().Error(errMsg)
			return
		}

		// adjuest fee
		fee := GetTxFee(redeemTx, feelevel_mid)
		redeemTxOut.Value = total - fee
	case tx_as_bs:
		var total int64
		for _, receiver := range receivers {
			total += receiver.Amount

			_, err = buildTxout(redeemTx, receiver, true)
			if err != nil {
				errMsg := fmt.Sprintf("buildMsgTx err:%+v", err)
				err = errors.New(errMsg)
				util.Logger().Error(errMsg)
				return
			}
		}

		var changeTxout *wire.TxOut
		changeTxout, err = buildTxout(redeemTx, &Target{Addr: changeReceiver}, false)
		if err != nil {
			errMsg := fmt.Sprintf("buildMsgTx err:%+v", err)
			err = errors.New(errMsg)
			util.Logger().Error(errMsg)
			return
		}

		outs := selectAS2BSOutputs(outputs, total)
		for _, out := range outs {
			balance := GetBalance(out)
			for _, o := range out {
				pkScript := buildTxin(redeemTx, o)
				PKScripts = append(PKScripts, pkScript)
			}

			// adjuest fee
			fee := GetTxFee(redeemTx, feelevel_mid)
			change := balance - total - fee
			if change > dirty_bitcoin {
				redeemTx.TxOut = append(redeemTx.TxOut, changeTxout)
				newFee := GetTxFee(redeemTx, feelevel_mid)
				newChange := balance - total - newFee
				if change >= dirty_bitcoin {
					changeTxout.Value = newChange
				} else {
					newOuts := util.RemoveIndex(redeemTx.TxOut, len(redeemTx.TxOut)-1)
					redeemTx.TxOut = newOuts
				}
				break
			} else if change < 0 {
				redeemTx.TxIn = nil
				PKScripts = nil
				continue
			} else if change >= 0 && change <= dirty_bitcoin {
				break
			}
		}
	default:
		err = errors.New("buildMsgTx: invalid tx type")
		return
	}
	return
}

func GetClientUnspent(wif string, netParams *chaincfg.Params) (client *txpool.TxpoolClient, outputs []*btcapi.UnspentOutput, from btcutil.Address, err error) {
	client = txpool.NewClient(netParams)
	from = BtcWif2Address(wif, netParams)
	outputs, err = client.ListUnspent(from)
	return
}

func MakeSignedTxAndBroastcast(t tx_type, client *txpool.TxpoolClient, receivers []*Target, wif string, from btcutil.Address, outputs []*btcapi.UnspentOutput) (txid string, err error) {
	redeemTx, PKScripts, err := buildMsgTx(t, receivers, from, outputs)
	if err != nil {
		errMsg := fmt.Sprintf("MakeSignedTxAndBroastcast err:%+v", err)
		util.Logger().Error(errMsg)
		err = errors.New(errMsg)
		return
	}

	signedTx, err := SignTx(wif, PKScripts, redeemTx)
	if err != nil {
		errMsg := fmt.Sprintf("MakeSignedTxAndBroastcast err:%+v", err)
		util.Logger().Error(errMsg)
		err = errors.New(errMsg)
		return
	}

	commitTxHash, err := client.BroadcastTx(signedTx)
	if err != nil {
		errMsg := fmt.Sprintf("MakeSignedTxAndBroastcast err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	txid = commitTxHash.String()
	return
}

// common transaction
// 1: common transfer
func A2BTrade(wif string, receiver *Target, netParams *chaincfg.Params) (txid string, err error) {
	client, outputs, from, err := GetClientUnspent(wif, netParams)
	if err != nil {
		errMsg := fmt.Sprintf("A2BTrade err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	var receivers []*Target
	receivers = append(receivers, receiver)
	txid, err = MakeSignedTxAndBroastcast(tx_a_b, client, receivers, wif, from, outputs)
	if err != nil {
		errMsg := fmt.Sprintf("AS2BTrade err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}
	return
}

// 2: found collect
func AS2BTrade(wif string, receiver btcutil.Address, netParams *chaincfg.Params) (txid string, err error) {
	client, outputs, _, err := GetClientUnspent(wif, netParams)
	if err != nil {
		errMsg := fmt.Sprintf("AS2BTrade err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	var receivers []*Target
	receivers = append(receivers, &Target{Addr: receiver})
	txid, err = MakeSignedTxAndBroastcast(tx_as_b, client, receivers, wif, nil, outputs)
	if err != nil {
		errMsg := fmt.Sprintf("AS2BTrade err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	return
}

// 3: pay for salary
func AS2BSTrade(wif string, receivers []*Target, netParams *chaincfg.Params) (txid string, err error) {
	client, outputs, from, err := GetClientUnspent(wif, netParams)
	if err != nil {
		errMsg := fmt.Sprintf("AS2BSTrade err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}
	var total int64 = 0
	for _, r := range receivers {
		total += r.Amount
	}

	balance := GetBalance(outputs)
	if balance < total {
		errMsg := "AS2BSTrade insufficient balance."
		util.Logger().Error(errMsg)
		return
	}

	txid, err = MakeSignedTxAndBroastcast(tx_as_bs, client, receivers, wif, from, outputs)
	if err != nil {
		errMsg := fmt.Sprintf("AS2BSTrade err:%+v", err)
		util.Logger().Error(errMsg)
		return
	}

	return
}

// multisig transaction
