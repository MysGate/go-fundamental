package trade

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MysGate/go-fundamental/util"
)

type feelevel int

const (
	feelevel_fast feelevel = iota
	feelevel_mid
	feelevel_slow
)

const (
	FEE_URL string = "https://mempool.space/api/v1/fees/recommended"
)

type FeeRate struct {
	FastestFee  int `json:"fastestFee"`
	HalfHourFee int `json:"halfHourFee"`
	HourFee     int `json:"hourFee"`
	EconomyFee  int `json:"economyFee"`
	MinimumFee  int `json:"minimumFee"`
}

type BTCFee struct {
	fee    map[feelevel]int64
	ticker *time.Ticker
	quit   chan bool
	locker *sync.RWMutex
}

var bf *BTCFee

func InitBtcFee() *BTCFee {
	bf = &BTCFee{
		fee:    make(map[feelevel]int64),
		quit:   make(chan bool, 1),
		locker: &sync.RWMutex{},
	}

	bf.updateFeeRate()
	go bf.triggerFetchFee()

	return bf
}

func GetBtcFee() *BTCFee {
	return bf
}

func (bf *BTCFee) Close() {
	bf.quit <- true
}

func (bf *BTCFee) triggerFetchFee() {
	bf.ticker = time.NewTicker(time.Hour)
	for {
		select {
		case <-bf.ticker.C:
			bf.updateFeeRate()
		case <-bf.quit:
			bf.ticker.Stop()
			return
		}
	}
}

func (bf *BTCFee) updateFeeRate() {
	headers := make(map[string]string)
	headers["Content-Type"] = " application/json"
	hc := util.GetHTTPClient()
	body, err := util.HTTPReq("GET", FEE_URL, hc, nil, headers)
	if err != nil {
		util.Logger().Error(fmt.Sprintf("updateFeeRate HTTPReq err:%+v", err))
		return
	}

	fr := FeeRate{}
	err = json.Unmarshal(body, &fr)
	if err != nil {
		util.Logger().Error(fmt.Sprintf("updateFeeRate HTTPReq err:%+v", err))
		return
	}

	bf.locker.Lock()
	defer bf.locker.Unlock()
	bf.fee[feelevel_fast] = int64(fr.FastestFee)
	bf.fee[feelevel_slow] = int64(fr.MinimumFee)
	bf.fee[feelevel_mid] = int64(float64(fr.HalfHourFee+fr.HourFee+fr.EconomyFee) / 3)
}

func (bf *BTCFee) GetFeeRate(level feelevel) int64 {
	bf.locker.RLock()
	defer bf.locker.RUnlock()
	return bf.fee[level]
}
