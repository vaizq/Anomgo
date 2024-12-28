package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID        uuid.UUID
	Hash      string
	CreatedAt time.Time
}

type TransactionMonel struct{}

func (m TransactionMonel) Create(ec db.ExecContext, txHash string) (*Transaction, error) {
	query := "INSERT INTO transactions (hash) VALUES($1) RETURNING id, created_at"

	tx := &Transaction{
		Hash: txHash,
	}

	err := ec.QueryRow(query, txHash).Scan(&tx.ID, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (m TransactionMonel) Delete(ec db.ExecContext, id uuid.UUID) error {
	query := "DELETE FROM transactions WHERE id = $1"
	_, err := ec.Exec(query, id)
	return err
}

func (m TransactionMonel) GetAll(ec db.ExecContext) ([]Transaction, error) {
	query := "SELECT id, hash, created_at FROM transactions"

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs := make([]Transaction, 0)
	for rows.Next() {
		tx := Transaction{}
		err := rows.Scan(&tx.ID, &tx.Hash, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (m TransactionMonel) GetAllAfter(ec db.ExecContext, tp time.Time) ([]Transaction, error) {
	query := "SELECT id, hash, created_at FROM transactions WHERE created_at < $1"

	rows, err := ec.Query(query, tp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs := make([]Transaction, 0)
	for rows.Next() {
		tx := Transaction{}
		err := rows.Scan(&tx.ID, &tx.Hash, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
