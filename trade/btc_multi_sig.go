package trade

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/MysGate/go-fundamental/chain/btc/btcapi"
	"github.com/MysGate/go-fundamental/util"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

type Receiver struct {
	Addr  btcutil.Address
	Ratio int
}

func DisAsembleScript(lockingScript string) (string, error) {
	// you can provide your locking script to dis asemble
	// lockingScript := "a914f63e2cbcc678236f683d267e7bb298ffdcd57b0487"
	script, err := hex.DecodeString(lockingScript)
	if err != nil {
		errmsg := fmt.Sprintf("DisAsembleScript: failed to makemultisig address err: %+v", err)
		util.Logger().Error(errmsg)
		return "", err
	}

	scriptStr, err := txscript.DisasmString(script)
	if err != nil {
		errmsg := fmt.Sprintf("DisAsembleScript: failed to makemultisig address err: %+v", err)
		util.Logger().Error(errmsg)
		return "", err
	}
	return scriptStr, nil
}

func BuildMultiSigScript(pubs []*btcutil.AddressPubKey, required int) (pkScript []byte, err error) {
	pkScript, err = txscript.MultiSigScript(pubs, required)
	if err != nil {
		errmsg := fmt.Sprintf("BuildMultiSigScript: failed to makemultisig address err: %+v", err)
		util.Logger().Error(errmsg)
		return
	}
	return
}

func BuildMultiSigP2SHAddr(pubs []*btcutil.AddressPubKey, required int, p *chaincfg.Params) (multiSigAddr string, pkScript []byte, err error) {
	pkScript, err = BuildMultiSigScript(pubs, required)
	if err != nil {
		errmsg := fmt.Sprintf("BuildMultiSigP2SHAddr: failed to makemultisig address err: %+v", err)
		util.Logger().Error(errmsg)
		return
	}

	scriptAddr, err := btcutil.NewAddressScriptHash(pkScript, p)
	if err != nil {
		errmsg := fmt.Sprintf("BuildMultiSigP2SHAddr: failed to make multisig address err: %+v", err)
		util.Logger().Error(errmsg)
		return
	}
	multiSigAddr = scriptAddr.String()
	return
}

func SignMultiSigTxin(wif string, redeemTx *wire.MsgTx, index int, redeemScript []byte) ([]byte, error) {
	decodedWif, err := btcutil.DecodeWIF(wif)
	if err != nil {
		errmsg := fmt.Sprintf("SignMultiSigTxin: failed to sig multisigtxin err: %+v", err)
		util.Logger().Error(errmsg)
		return nil, err
	}
	sig, err := txscript.RawTxInSignature(redeemTx, index, redeemScript, txscript.SigHashAll, decodedWif.PrivKey)
	if err != nil {
		errmsg := fmt.Sprintf("SignMultiSigTxin: failed to sig multisigtxin err: %+v", err)
		util.Logger().Error(errmsg)
		return nil, err
	}
	return sig, nil
}

func SpendMultiSig(sigs [][]byte, required int, receivers []*Receiver, outputs []*btcapi.UnspentOutput, redeemScript []byte) (*wire.MsgTx, error) {
	if len(sigs) == 0 {
		errmsg := "SpendMultiSig: sig err"
		err := errors.New(errmsg)
		util.Logger().Error(errmsg)
		return nil, err
	}

	var totalRatio int
	for _, r := range receivers {
		totalRatio += r.Ratio
	}

	if totalRatio != 100 {
		errmsg := "SpendMultiSig: ratio err"
		err := errors.New(errmsg)
		util.Logger().Error(errmsg)
		return nil, err
	}

	var totalValue int64
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	for _, output := range outputs {
		totalValue += output.Output.Value

		outPoint := wire.NewOutPoint(&output.Outpoint.Hash, output.Outpoint.Index)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		redeemTx.AddTxIn(txIn)
	}

	for _, r := range receivers {
		destAddrByte, err := txscript.PayToAddrScript(r.Addr)
		if err != nil {
			errmsg := fmt.Sprintf("SpendMultiSig: failed to make multisig address err: %+v", err)
			util.Logger().Error(errmsg)
			return nil, err
		}

		redeemTxOut := wire.NewTxOut(0, destAddrByte)
		redeemTx.AddTxOut(redeemTxOut)
	}

	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_0)
	for _, sig := range sigs {
		signature.AddData(sig)
	}

	signatureScript, err := signature.Script()
	if err != nil {
		errmsg := fmt.Sprintf("SpendMultiSig: failed to build script err: %+v", err)
		util.Logger().Error(errmsg)
		return nil, err
	}

	for _, txin := range redeemTx.TxIn {
		txin.SignatureScript = signatureScript
	}

	// calculate fee
	fee := GetTxFee(redeemTx, feelevel_mid)
	totalValue = totalValue - fee
	for idx, r := range receivers {
		value := int64(float64(totalValue) * (float64(r.Ratio) / 100))
		redeemTx.TxOut[idx].Value = value
	}

	return redeemTx, nil
}
