package payment

import (
	"LuomuTori/internal/log"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync/atomic"
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"
)

var url = "https://min-api.cryptocompare.com/data/price?fsym=XMR&tsyms=USD,EUR"

var latestXMRPrice atomic.Int32

type prices struct {
	USD float64
	EUR float64
}

func updateXMRPrice() (*prices, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid response status: %s\n", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	prices := &prices{}
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, err
	}

	return prices, nil
}

func UpdateXMRPrice() error {
	var err error = nil

	for i := 0; i < 5; i++ {
		var prices *prices = nil
		if prices, err = updateXMRPrice(); err == nil {
			latestXMRPrice.Store(int32(prices.EUR))
			return nil
		}
		log.Error.Printf("Failed to update XMR price (try %d): %s\n", i+1, err.Error())
		time.Sleep(time.Minute)
	}

	return err
}

// How much 1XMR is in Fiat (EUR)
func XMRPrice() float64 {
	return float64(latestXMRPrice.Load())
}

func Fiat2XMR(fiat float64) uint64 {
	price := XMRPrice()
	return uint64(XMRf * math.Abs(fiat/price))
}

func XMR2Fiat(xmr uint64) float64 {
	return walletrpc.XMRToFloat64(xmr) * XMRPrice()
}

var XMR2Float = walletrpc.XMRToFloat64
var XMR2Decimal = walletrpc.XMRToDecimal

const XMR uint64 = 1e12
const XMRf float64 = float64(XMR)

func TakeCut(amount uint64) uint64 {
	return uint64(float64(amount) * 0.95)
}
