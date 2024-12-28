package main

import (
	"testing"
)

func TestParser(t *testing.T) {
	page := `<p>{{T "Order" $.Lang}}</p>`
	phrases := parsePhrases(page)

	if len(phrases) != 1 {
		t.Fatal("Error")
	}
	if phrases[0] != "Order" {
		t.Fatalf("Should be Order but is %s\n", phrases[0])
	}
}
