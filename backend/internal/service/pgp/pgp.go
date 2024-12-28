package pgp

import (
	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

func PublicKeyIsValid(key string) bool {
	if _, err := crypto.NewKeyFromArmored(key); err == nil {
		return true
	} else {
		return false
	}
}

func SignatureIsValid(pubKey string, signature string) bool {
	key, err := crypto.NewKeyFromArmored(pubKey)
	if err != nil {
		return false
	}
	pgp := crypto.PGP()
	verifier, err := pgp.Verify().VerificationKey(key).New()
	if err != nil {
		return false
	}
	verifyResult, err := verifier.VerifyCleartext([]byte(signature))
	if err != nil {
		return false
	}
	if sigErr := verifyResult.SignatureError(); sigErr != nil {
		return false
	}

	return true
}

func EncryptMessage(pubkey, message string) (string, error) {
	publicKey, err := crypto.NewKeyFromArmored(pubkey)

	pgp := crypto.PGP()
	// Encrypt plaintext message using a public key
	encHandle, err := pgp.Encryption().Recipient(publicKey).New()
	if err != nil {
		return "", err
	}

	encrypted, err := encHandle.Encrypt([]byte(message))
	if err != nil {
		return "", err
	}

	return encrypted.Armor()
}
