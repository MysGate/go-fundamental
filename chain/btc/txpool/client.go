package txpool

import (
	"io"
	"log"

	"github.com/MysGate/go-fundamental/chain/btc/btcapi"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

type TxpoolClient struct {
	baseURL string
}

func NewClient(netParams *chaincfg.Params) *TxpoolClient {
	baseURL := ""
	if netParams.Net == wire.MainNet {
		baseURL = "https://mempool.space/api"
	} else if netParams.Net == wire.TestNet3 {
		baseURL = "https://mempool.space/testnet/api"
	} else if netParams.Net == chaincfg.SigNetParams.Net {
		baseURL = "https://mempool.space/signet/api"
	} else {
		log.Fatal("mempool don't support other netParams")
	}
	return &TxpoolClient{
		baseURL: baseURL,
	}
}

func (c *TxpoolClient) request(method, subPath string, requestBody io.Reader) ([]byte, error) {
	return btcapi.Request(method, c.baseURL, subPath, requestBody)
}

var _ btcapi.BTCAPIClient = (*TxpoolClient)(nil)
