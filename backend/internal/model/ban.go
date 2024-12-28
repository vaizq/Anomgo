package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type Ban struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
}

type BanModel struct{}

func (um BanModel) Create(ec db.ExecContext, userID uuid.UUID) (*Ban, error) {
	b := &Ban{
		UserID: userID,
	}

	if err := ec.QueryRow("INSERT INTO bans (user_id) VALUES($1) RETURNING id, created_at", userID).Scan(&b.ID, &b.CreatedAt); err != nil {
		return nil, err
	}

	return b, nil
}

func (um BanModel) GetForUser(ec db.ExecContext, userID uuid.UUID) (*Ban, error) {
	b := &Ban{
		UserID: userID,
	}

	if err := ec.QueryRow("SELECT id, created_at FROM bans WHERE user_id = $1", userID).Scan(&b.ID, &b.CreatedAt); err != nil {
		return nil, err
	}

	return b, nil
}

func (um BanModel) Delete(ec db.ExecContext, id uuid.UUID) error {
	_, err := ec.Exec("DELETE FROM bans WHERE id = $1", id)
	return err
}
