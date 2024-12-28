package payment

import (
	"gitlab.com/moneropay/go-monero/walletrpc"
	"math"
	"testing"
)

func TestConversions(t *testing.T) {
	if err := UpdateXMRPrice(); err != nil {
		t.Fatalf("Failed to update XMR price\n")
	}

	if err := math.Abs(walletrpc.XMRToFloat64(Fiat2XMR(XMRPrice())) - 1.0); err != 0 {
		t.Fatalf("err is not 0, but %f\n", err)
	}

	if price := XMRPrice(); price != XMR2Fiat(1e12) {
		t.Fatalf("Price of 1e12 piconeros should equal price of one XMR %f\n", price)
	}

	if TakeCut(100) != 95 {
		t.Fatalf("Failed to take cut\n")
	}
}
