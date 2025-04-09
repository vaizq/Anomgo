package auth

import (
	mydb "LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"LuomuTori/internal/service/payment"
	"LuomuTori/internal/service/pgp"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameAlreadyRegistered = errors.New("Username unavailable")
	ErrInvalidCredentials        = errors.New("Invalid credentials")
	ErrInvalidPassword           = errors.New("Invalid password")
	ErrInvalidPGPKey             = errors.New("invalid PGP public key")
	ErrInvalidSignature          = errors.New("Invalid signature")
	ErrAccountIsBanned           = errors.New("Account is banned")
)

func Register(db *sql.DB, username, password string, pgpKey string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if pgpKey != "" {
		if !pgp.PublicKeyIsValid(pgpKey) {
			return nil, ErrInvalidPGPKey
		}
	}

	u, err := model.M.User.Create(tx, username, hash, pgpKey)
	if err != nil {
		if mydb.ErrCode(err) == mydb.ErrCodeUniqueViolation {
			return nil, ErrUsernameAlreadyRegistered
		}
		return nil, err
	}

	invoice, err := payment.CreateInvoiceForDeposits(u.ID)
	if err != nil {
		return nil, err
	}

	_, err = model.M.Wallet.Create(tx, u.ID, invoice.Address)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return u, nil
}

func Authenticate(db *sql.DB, username, password string) (*model.User, error) {
	user, err := model.M.User.GetWithName(db, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if ban, _ := model.M.Ban.GetForUser(db, user.ID); ban != nil {
		return nil, ErrAccountIsBanned
	}

	hash, err := model.M.User.GetHashedPassword(db, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// Generates random string of letters and numbers of length n
func Generate2FAToken(n int) (string, error) {
	rb := make([]byte, n)
	if _, err := rand.Read(rb); err != nil {
		return "", err
	}

	token := fmt.Sprintf("%x", sha256.Sum256(rb))[0:n]
	return token, nil
}

func Authenticate2FA(db *sql.DB, username, password, signature string) (*model.User, error) {
	user, err := model.M.User.GetWithName(db, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	hash, err := model.M.User.GetHashedPassword(db, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PgpKey != nil {
		if signature == "" {
			return nil, ErrInvalidSignature
		}

		if !pgp.SignatureIsValid(*user.PgpKey, signature) {
			return nil, ErrInvalidSignature
		}
	}

	return user, nil
}

func ChangePassword(db *sql.DB, username string, oldPassword string, newPassword string) (*model.User, error) {
	user, err := model.M.User.GetWithName(db, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	hash, err := model.M.User.GetHashedPassword(db, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(oldPassword))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err = model.M.User.UpdatePasswordHash(db, username, hash, newHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func Enable2FA(db *sql.DB, userID uuid.UUID, pgpKey string) (*model.User, error) {
	if !pgp.PublicKeyIsValid(pgpKey) {
		return nil, ErrInvalidPGPKey
	}
	return model.M.User.UpdatePgpKey(db, userID, pgpKey)
}
