package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
)

type Wallet struct {
	ID      uuid.UUID
	Address string
	Balance uint64
	UserID  uuid.UUID
}

type WalletModel struct{}

func (m WalletModel) Create(ec db.ExecContext, userID uuid.UUID, address string) (*Wallet, error) {
	query := "INSERT INTO wallets (address, user_id) VALUES($1, $2) RETURNING id"

	wallet := &Wallet{
		Address: address,
		Balance: 0,
		UserID:  userID,
	}

	err := ec.QueryRow(query, address, userID).Scan(&wallet.ID)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (m WalletModel) Get(ec db.ExecContext, id uuid.UUID) (*Wallet, error) {
	query := "SELECT address, balance, user_id FROM wallets WHERE id = $1"

	wallet := &Wallet{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&wallet.Address, &wallet.Balance, &wallet.UserID)
	if err != nil {
		return nil, err
	}
	return wallet, nil

}

func (m WalletModel) GetAll(ec db.ExecContext) ([]Wallet, error) {
	query := "SELECT id, address, balance, user_id FROM wallets"

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wallets := make([]Wallet, 0)

	for rows.Next() {
		w := Wallet{}
		err := rows.Scan(&w.ID, &w.Address, &w.Balance, &w.UserID)
		if err != nil {
			return nil, err
		}

		wallets = append(wallets, w)
	}

	return wallets, nil
}

func (m WalletModel) GetForUser(ec db.ExecContext, userID uuid.UUID) (*Wallet, error) {
	query := `
	SELECT wallets.id, wallets.address, wallets.balance
	FROM wallets 
	JOIN users ON wallets.user_id = users.id
	WHERE users.id = $1
	`

	wallet := &Wallet{
		UserID: userID,
	}

	err := ec.QueryRow(query, userID).Scan(&wallet.ID, &wallet.Address, &wallet.Balance)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (m WalletModel) SetBalance(ec db.ExecContext, id uuid.UUID, balance uint64) (*Wallet, error) {
	query := `
	UPDATE wallets 
	SET balance = $2 
	WHERE id = $1 
	RETURNING address, balance, user_id
	`
	wallet := &Wallet{
		ID: id,
	}

	err := ec.QueryRow(query, id, balance).Scan(&wallet.Address, &wallet.Balance, &wallet.UserID)
	if err != nil {
		return nil, err
	}
	return wallet, nil

}

func (m WalletModel) AddBalance(ec db.ExecContext, id uuid.UUID, amount uint64) (*Wallet, error) {
	query := `
	UPDATE wallets 
	SET balance = balance + $2 
	WHERE id = $1 
	RETURNING address, balance, user_id
	`
	wallet := &Wallet{
		ID: id,
	}

	err := ec.QueryRow(query, id, amount).Scan(&wallet.Address, &wallet.Balance, &wallet.UserID)
	if err != nil {
		return nil, err
	}
	return wallet, nil

}

func (m WalletModel) ReduceBalance(ec db.ExecContext, id uuid.UUID, amount uint64) (*Wallet, error) {
	query := `
	UPDATE wallets 
	SET balance = balance - $2 
	WHERE id = $1 AND balance >= $2
	RETURNING address, balance, user_id
	`
	wallet := &Wallet{
		ID: id,
	}

	err := ec.QueryRow(query, id, amount).Scan(&wallet.Address, &wallet.Balance, &wallet.UserID)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (m WalletModel) UpdateAddress(ec db.ExecContext, id uuid.UUID, address string) (*Wallet, error) {
	query := `
	UPDATE wallets 
	SET address = $2 
	WHERE id = $1
	RETURNING balance, user_id
	`
	wallet := &Wallet{
		ID:      id,
		Address: address,
	}

	err := ec.QueryRow(query, id, address).Scan(&wallet.Balance, &wallet.UserID)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
