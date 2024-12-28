package validate

import (
	"regexp"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

func (v *Validator) CheckField(ok bool, field string, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if !ok {
		v.FieldErrors[field] = message
	}
}

func (v *Validator) SetError(message string) {
	if v.NonFieldErrors == nil {
		v.NonFieldErrors = make([]string, 0)
	}

	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

var emailRegexp, _ = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func ValidEmail(email string) bool {
	return emailRegexp.Match([]byte(email))
}

func ValidXMRAddress(address string) bool {
	return len(address) == 95
}

func AtleastNRunes(s string, n int) bool {
	return utf8.RuneCountInString(s) >= n
}

func AtmostNRunes(s string, n int) bool {
	return utf8.RuneCountInString(s) <= n
}
