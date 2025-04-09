package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID
	Username  string
	PgpKey    *string
	PrevLogin time.Time
	CreatedAt time.Time
}

type UserModel struct{}

func (um UserModel) Create(ec db.ExecContext, username string, passwordHash []byte, pgpKey string) (*User, error) {
	u := &User{
		Username: username,
	}

	var err error
	if pgpKey == "" {
		err = ec.QueryRow("INSERT INTO users (username, hashed_password) VALUES($1, $2) RETURNING id, prev_login, created_at",
			username, passwordHash).Scan(&u.ID, &u.PrevLogin, &u.CreatedAt)
	} else {
		err = ec.QueryRow("INSERT INTO users (username, hashed_password, pgp_key) VALUES($1, $2, $3) RETURNING id, pgp_key, prev_login, created_at",
			username, passwordHash, pgpKey).Scan(&u.ID, &u.PgpKey, &u.PrevLogin, &u.CreatedAt)
	}

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (um UserModel) UpdatePasswordHash(ec db.ExecContext, username string, oldHash []byte, newHash []byte) (*User, error) {
	query := `
		UPDATE users 
		SET hashed_password = $3 
		WHERE username = $1 AND hashed_password = $2
		RETURNING id, pgp_key, prev_login, created_at
	`

	user := &User{
		Username: username,
	}

	err := ec.QueryRow(query, username, oldHash, newHash).Scan(&user.ID, &user.PgpKey, &user.PrevLogin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (um UserModel) UpdatePgpKey(ec db.ExecContext, id uuid.UUID, pgpKey string) (*User, error) {
	query := `
		UPDATE users 
		SET pgp_key = $2 
		WHERE id = $1
		RETURNING username, pgp_key, prev_login, created_at
	`

	user := &User{
		ID: id,
	}

	err := ec.QueryRow(query, id, pgpKey).Scan(&user.Username, &user.PgpKey, &user.PrevLogin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (um UserModel) UpdatePrevLogin(ec db.ExecContext, id uuid.UUID) (*User, error) {
	query := `
		UPDATE users 
		SET prev_login = NOW() 
		WHERE id = $1
		RETURNING username, pgp_key, prev_login, created_at`

	user := &User{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&user.Username, &user.PgpKey, &user.PrevLogin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (um UserModel) Get(ec db.ExecContext, id uuid.UUID) (*User, error) {
	query := "SELECT username, pgp_key, prev_login, created_at FROM users WHERE id = $1"

	u := &User{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&u.Username, &u.PgpKey, &u.PrevLogin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (um UserModel) GetWithName(ec db.ExecContext, username string) (*User, error) {
	query := "SELECT id, pgp_key, prev_login, created_at FROM users WHERE username = $1"

	u := &User{
		Username: username,
	}

	err := ec.QueryRow(query, username).Scan(&u.ID, &u.PgpKey, &u.PrevLogin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (um UserModel) GetHashedPassword(ec db.ExecContext, id uuid.UUID) ([]byte, error) {
	query := "SELECT hashed_password FROM users WHERE id = $1"
	hashedPassword := make([]byte, 0)

	err := ec.QueryRow(query, id).Scan(&hashedPassword)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}

func (m UserModel) IsVendor(ec db.ExecContext, id uuid.UUID) bool {
	query := "SELECT 1 FROM vendor_pledges WHERE user_id = $1"
	var i int = 0
	err := ec.QueryRow(query, id).Scan(&i)
	return err == nil && i == 1
}
