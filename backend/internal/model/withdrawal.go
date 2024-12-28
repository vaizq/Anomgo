package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type WithdrawalStatus string

const (
	WithdrawalPending    WithdrawalStatus = "pending"
	WithdrawalProcessing WithdrawalStatus = "processing"
)

type Withdrawal struct {
	ID          uuid.UUID
	Amount      uint64
	DestAddress string
	Status      WithdrawalStatus
	CreatedAt   time.Time
}

type WithdrawalModel struct{}

func (m WithdrawalModel) Create(ec db.ExecContext, destAddress string, amount uint64, status WithdrawalStatus) (*Withdrawal, error) {
	query := "INSERT INTO withdrawals (amount, dest_address, status) VALUES($1, $2, $3) RETURNING id, created_at"

	withdrawal := &Withdrawal{
		Amount:      0,
		DestAddress: destAddress,
		Status:      status,
	}

	err := ec.QueryRow(query, amount, destAddress, status).Scan(&withdrawal.ID, &withdrawal.CreatedAt)
	if err != nil {
		return nil, err
	}
	return withdrawal, nil

}

func (m WithdrawalModel) Delete(ec db.ExecContext, id uuid.UUID) error {
	query := "DELETE FROM withdrawals WHERE id = $1"
	_, err := ec.Exec(query, id)
	return err
}

func (m WithdrawalModel) GetAllWithStatus(ec db.ExecContext, status WithdrawalStatus) ([]Withdrawal, error) {
	query := "SELECT * FROM withdrawals WHERE status = $1"

	rows, err := ec.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ws := []Withdrawal{}

	for rows.Next() {
		w := Withdrawal{}
		if err := rows.Scan(&w.ID, &w.Amount, &w.DestAddress, &w.Status, &w.CreatedAt); err != nil {
			return nil, err
		}

		ws = append(ws, w)
	}

	return ws, nil
}

func (m WithdrawalModel) UpdateStatus(ec db.ExecContext, id uuid.UUID, status WithdrawalStatus) (*Withdrawal, error) {
	query := "UPDATE withdrawals SET status = $2 WHERE id = $1 RETURNING amount, dest_address, created_at"

	w := &Withdrawal{
		ID:     id,
		Status: status,
	}

	if err := ec.QueryRow(query, id, status).Scan(&w.Amount, &w.DestAddress, &w.CreatedAt); err != nil {
		return nil, err
	}

	return w, nil
}
