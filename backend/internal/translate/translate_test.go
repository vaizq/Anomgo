package translate

import (
	"testing"
)

func TestTranslations(t *testing.T) {
	w := "register"
	if r := T(w, Fi); r != "rekisteröidy" {
		t.Fatal(w, r)
	}
	w = "Register"
	if r := T(w, Fi); r != "Rekisteröidy" {
		t.Fatal(w, r)
	}
}
