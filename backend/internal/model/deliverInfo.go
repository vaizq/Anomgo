package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type DeliveryInfo struct {
	ID        uuid.UUID
	Info      string
	OrderID   uuid.UUID
	CreatedAt time.Time
}

type DeliveryInfoModel struct{}

func (pm DeliveryInfoModel) Create(ec db.ExecContext, info string, orderID uuid.UUID) (*DeliveryInfo, error) {
	query := "INSERT INTO delivery_infos (info, order_id) VALUES ($1, $2) RETURNING id, created_at"

	di := &DeliveryInfo{
		Info:    info,
		OrderID: orderID,
	}

	if err := ec.QueryRow(query, info, orderID).Scan(&di.ID, &di.CreatedAt); err != nil {
		return nil, err
	}

	return di, nil
}

func (pm DeliveryInfoModel) GetForOrder(ec db.ExecContext, orderID uuid.UUID) (*DeliveryInfo, error) {
	query := "SELECT id, info, created_at FROM delivery_infos WHERE order_id=$1"

	di := &DeliveryInfo{
		OrderID: orderID,
	}

	if err := ec.QueryRow(query, orderID).Scan(&di.ID, &di.Info, &di.CreatedAt); err != nil {
		return nil, err
	}

	return di, nil
}
