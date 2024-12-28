package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
)

type Order struct {
	Order          *model.Order
	Product        *model.Product
	Price          *model.Price
	DeliveryMethod *model.DeliveryMethod
	Invoice        *model.Invoice
	DeliveryInfo   *model.DeliveryInfo
	DeclineReason  *model.DeclineReason
}

type OrderView struct{}

func (ov OrderView) Get(ec db.ExecContext, orderID uuid.UUID) (*Order, error) {
	order, err := model.M.Order.Get(ec, orderID)
	if err != nil {
		return nil, err
	}

	price, err := model.M.Price.Get(ec, order.PriceID)
	if err != nil {
		return nil, err
	}

	dm, err := model.M.DeliveryMethod.GetForOrder(ec, orderID)
	if err != nil {
		return nil, err
	}

	product, err := model.M.Product.Get(ec, price.ProductID)
	if err != nil {
		return nil, err
	}

	invoice, err := model.M.Invoice.GetWithOrderID(ec, order.ID)

	view := &Order{
		Order:          order,
		Product:        product,
		Price:          price,
		DeliveryMethod: dm,
		Invoice:        invoice,
	}

	switch {
	case order.Status != model.StatusPaid && order.Status != model.StatusDeclined:
		di, err := model.M.DeliveryInfo.GetForOrder(ec, order.ID)
		if err != nil {
			return nil, err
		}
		view.DeliveryInfo = di
	case order.Status == model.StatusDeclined:
		dr, err := model.M.DeclineReason.GetForOrder(ec, order.ID)
		if err != nil {
			return nil, err
		}
		view.DeclineReason = dr
	}

	return view, nil
}

func (ov OrderView) GetAllForCustomer(ec db.ExecContext, customerID uuid.UUID) ([]Order, error) {
	query := `
	SELECT orders.id, orders.status, orders.details, orders.price_id, orders.customer_id, orders.created_at,
		products.id, products.title, products.description, products.image_filename, products.vendor_id,
		prices.id, prices.quantity, prices.price, prices.product_id,
		dm.id, dm.description, dm.price, dm.product_id,
		invoices.id, invoices.address, invoices.order_id, invoices.xmr_price, invoices.created_at
	FROM users
	JOIN orders ON orders.customer_id = users.id
	JOIN invoices ON invoices.order_id = orders.id
	JOIN prices ON prices.id = orders.price_id
	JOIN delivery_methods AS dm ON dm.id = orders.delivery_method_id
	JOIN products ON products.id = prices.product_id
	WHERE users.id = $1
	ORDER BY orders.created_at DESC
	`

	rows, err := ec.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		order := &model.Order{}
		product := &model.Product{}
		price := &model.Price{}
		dm := &model.DeliveryMethod{}
		invoice := &model.Invoice{}

		err = rows.Scan(
			&order.ID, &order.Status, &order.Details, &order.PriceID, &order.CustomerID, &order.CreatedAt,
			&product.ID, &product.Title, &product.Description, &product.ImageFilename, &product.VendorID,
			&price.ID, &price.Quantity, &price.Price, &price.ProductID,
			&dm.ID, &dm.Description, &dm.Price, &dm.ProductID,
			&invoice.ID, &invoice.Address, &invoice.OrderID, &invoice.XMRPrice, &invoice.CreatedAt)
		if err != nil {
			return nil, err
		}

		view := Order{
			Order:          order,
			Product:        product,
			Price:          price,
			DeliveryMethod: dm,
			Invoice:        invoice,
		}

		switch order.Status {
		case model.StatusDelivered:
			di, err := model.M.DeliveryInfo.GetForOrder(ec, order.ID)
			if err != nil {
				return nil, err
			}
			view.DeliveryInfo = di
		case model.StatusDeclined:
			dr, err := model.M.DeclineReason.GetForOrder(ec, order.ID)
			if err != nil {
				return nil, err
			}
			view.DeclineReason = dr
		}

		orders = append(orders, view)
	}

	return orders, nil
}

func (ov OrderView) GetAllForVendor(ec db.ExecContext, vendorID uuid.UUID) ([]Order, error) {
	query := `
	SELECT orders.id, orders.status, orders.details, orders.price_id, orders.customer_id, orders.created_at,
		products.id, products.title, products.description, products.image_filename, products.vendor_id,
		prices.id, prices.quantity, prices.price, prices.product_id,
		dm.id, dm.description, dm.price, dm.product_id
	FROM products
	JOIN prices ON prices.product_id = products.id 
	JOIN orders ON orders.price_id = prices.id
	JOIN delivery_methods AS dm ON dm.id = orders.delivery_method_id
	WHERE products.vendor_id = $1
	ORDER BY orders.created_at DESC
	`

	rows, err := ec.Query(query, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		order := &model.Order{}
		product := &model.Product{}
		price := &model.Price{}
		dm := &model.DeliveryMethod{}

		err = rows.Scan(
			&order.ID, &order.Status, &order.Details, &order.PriceID, &order.CustomerID, &order.CreatedAt,
			&product.ID, &product.Title, &product.Description, &product.ImageFilename, &product.VendorID,
			&price.ID, &price.Quantity, &price.Price, &price.ProductID,
			&dm.ID, &dm.Description, &dm.Price, &dm.ProductID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, Order{
			Order:          order,
			Product:        product,
			Price:          price,
			DeliveryMethod: dm,
		})
	}

	return orders, nil
}
