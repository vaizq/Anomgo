package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID
	Title         string
	Description   string
	ImageFilename string
	VendorID      uuid.UUID
}

type ProductModel struct{}

func (pm *ProductModel) Create(ec db.ExecContext, title string, description string, imageFilename string, vendorID uuid.UUID) (*Product, error) {
	query := "INSERT INTO products (title, description, image_filename, vendor_id) VALUES ($1, $2, $3, $4) RETURNING id"
	p := &Product{
		Title:         title,
		Description:   description,
		ImageFilename: imageFilename,
		VendorID:      vendorID,
	}

	err := ec.QueryRow(query, title, description, imageFilename, vendorID).Scan(&p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (pm *ProductModel) Get(ec db.ExecContext, id uuid.UUID) (*Product, error) {
	query := "SELECT title, description, image_filename, vendor_id FROM products WHERE id = $1"

	p := &Product{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&p.Title, &p.Description, &p.ImageFilename, &p.VendorID)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (pm *ProductModel) GetAll(ec db.ExecContext) ([]Product, error) {
	query := "SELECT id, title, description, image_filename, vendor_id FROM products WHERE deleted_at IS NULL"

	products := make([]Product, 0)

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.ImageFilename, &p.VendorID)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (pm *ProductModel) Delete(ec db.ExecContext, id uuid.UUID) error {
	query := "UPDATE products SET deleted_at = NOW() WHERE id=$1"
	_, err := ec.Exec(query, id)
	return err
}
