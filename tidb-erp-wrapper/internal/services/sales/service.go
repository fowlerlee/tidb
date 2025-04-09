package sales

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/db"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
)

type Service struct {
	db *db.DBHandler
}

func NewService(db *db.DBHandler) *Service {
	return &Service{db: db}
}

func (s *Service) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO customers (
			name, code, tax_id, address, contact_info,
			payment_terms, credit_limit, currency_code,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		customer.Name,
		customer.Code,
		customer.TaxID,
		customer.Address,
		customer.ContactInfo,
		customer.PaymentTerms,
		customer.CreditLimit,
		customer.CurrencyCode,
		customer.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating customer: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	customer.ID = id
	return nil
}

func (s *Service) CreatePriceList(ctx context.Context, priceList *models.PriceList, items []models.PriceListItem) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO price_lists (
			name, description, currency_code, start_date,
			end_date, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		priceList.Name,
		priceList.Description,
		priceList.CurrencyCode,
		priceList.StartDate,
		priceList.EndDate,
		priceList.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating price list: %v", err)
	}

	priceListID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting price list ID: %v", err)
	}
	priceList.ID = priceListID

	for _, item := range items {
		query := `
			INSERT INTO price_list_items (
				price_list_id, product_code, unit_price,
				min_quantity, discount_percentage,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			priceListID,
			item.ProductCode,
			item.UnitPrice,
			item.MinQuantity,
			item.DiscountPercent,
		)
		if err != nil {
			return fmt.Errorf("error creating price list item: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Service) CreateSalesOrder(ctx context.Context, order *models.SalesOrder, items []models.SalesOrderItem) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Validate customer credit limit
	var customerCreditLimit, customerCurrentOrders float64
	err = tx.QueryRowContext(ctx, "SELECT credit_limit FROM customers WHERE id = ?", order.CustomerID).Scan(&customerCreditLimit)
	if err != nil {
		return fmt.Errorf("error getting customer credit limit: %v", err)
	}

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0) 
		FROM sales_orders 
		WHERE customer_id = ? AND status IN ('open', 'processing')
	`, order.CustomerID).Scan(&customerCurrentOrders)
	if err != nil {
		return fmt.Errorf("error getting customer current orders: %v", err)
	}

	if customerCurrentOrders+order.TotalAmount > customerCreditLimit {
		return errors.New("order exceeds customer credit limit")
	}

	// Insert sales order
	query := `
		INSERT INTO sales_orders (
			order_number, customer_id, price_list_id,
			order_date, requested_delivery_date, status,
			total_amount, tax_amount, currency_code,
			payment_terms, sales_person, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		order.OrderNumber,
		order.CustomerID,
		order.PriceListID,
		order.OrderDate,
		order.RequestedDeliveryDate,
		order.Status,
		order.TotalAmount,
		order.TaxAmount,
		order.CurrencyCode,
		order.PaymentTerms,
		order.SalesPerson,
		order.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating sales order: %v", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting sales order ID: %v", err)
	}
	order.ID = orderID

	// Insert sales order items
	for _, item := range items {
		query := `
			INSERT INTO sales_order_items (
				sales_order_id, product_code, description,
				quantity, unit_price, discount_percentage,
				total_price, tax_rate, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			orderID,
			item.ProductCode,
			item.Description,
			item.Quantity,
			item.UnitPrice,
			item.DiscountPercent,
			item.TotalPrice,
			item.TaxRate,
		)
		if err != nil {
			return fmt.Errorf("error creating sales order item: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Service) CreateDelivery(ctx context.Context, delivery *models.Delivery, items []models.DeliveryItem) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Validate sales order status
	var orderStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM sales_orders WHERE id = ?", delivery.SalesOrderID).Scan(&orderStatus)
	if err != nil {
		return fmt.Errorf("error getting sales order status: %v", err)
	}
	if orderStatus != "confirmed" && orderStatus != "processing" {
		return errors.New("cannot create delivery for unconfirmed or completed sales order")
	}

	// Insert delivery
	query := `
		INSERT INTO deliveries (
			delivery_number, sales_order_id, delivery_date,
			status, shipping_address, tracking_number,
			carrier, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		delivery.DeliveryNumber,
		delivery.SalesOrderID,
		delivery.DeliveryDate,
		delivery.Status,
		delivery.ShippingAddress,
		delivery.TrackingNumber,
		delivery.Carrier,
		delivery.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating delivery: %v", err)
	}

	deliveryID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting delivery ID: %v", err)
	}
	delivery.ID = deliveryID

	// Insert delivery items and update inventory
	for _, item := range items {
		// Get product ID from product code
		var productID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM products WHERE code = ?", item.BatchNumber).Scan(&productID)
		if err != nil {
			return fmt.Errorf("error getting product ID: %v", err)
		}

		// Insert delivery item
		query := `
			INSERT INTO delivery_items (
				delivery_id, sales_order_item_id, quantity,
				batch_number, serial_numbers, notes,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			deliveryID,
			item.SalesOrderItemID,
			item.Quantity,
			item.BatchNumber,
			item.SerialNumbers,
			item.Notes,
		)
		if err != nil {
			return fmt.Errorf("error creating delivery item: %v", err)
		}

		// Create inventory transaction for the delivery
		query = `
			INSERT INTO inventory_transactions (
				product_id, warehouse_id, storage_location_id,
				transaction_type, reference_type, reference_id,
				quantity, transaction_date, created_at, updated_at
			) VALUES (?, ?, ?, 'outbound', 'delivery', ?, ?, NOW(), NOW(), NOW())
		`
		_, err = tx.ExecContext(ctx, query,
			productID,
			1, // Default warehouse ID, should be configurable
			1, // Default storage location ID, should be configurable
			deliveryID,
			-item.Quantity, // Negative quantity for outbound transaction
		)
		if err != nil {
			return fmt.Errorf("error creating inventory transaction: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Service) GetSalesOrderByID(ctx context.Context, id int64) (*models.SalesOrder, []models.SalesOrderItem, error) {
	order := &models.SalesOrder{}
	query := `
		SELECT id, order_number, customer_id, price_list_id,
			   order_date, requested_delivery_date, status,
			   total_amount, tax_amount, currency_code,
			   payment_terms, sales_person, notes,
			   created_at, updated_at
		FROM sales_orders
		WHERE id = ?
	`
	err := s.db.DB().QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.PriceListID,
		&order.OrderDate,
		&order.RequestedDeliveryDate,
		&order.Status,
		&order.TotalAmount,
		&order.TaxAmount,
		&order.CurrencyCode,
		&order.PaymentTerms,
		&order.SalesPerson,
		&order.Notes,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("sales order not found")
		}
		return nil, nil, fmt.Errorf("error querying sales order: %v", err)
	}

	items := []models.SalesOrderItem{}
	query = `
		SELECT id, product_code, description, quantity,
			   unit_price, discount_percentage, total_price,
			   tax_rate, created_at, updated_at
		FROM sales_order_items
		WHERE sales_order_id = ?
	`
	rows, err := s.db.DB().QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying sales order items: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.SalesOrderItem
		err := rows.Scan(
			&item.ID,
			&item.ProductCode,
			&item.Description,
			&item.Quantity,
			&item.UnitPrice,
			&item.DiscountPercent,
			&item.TotalPrice,
			&item.TaxRate,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning sales order item: %v", err)
		}
		item.SalesOrderID = id
		items = append(items, item)
	}

	return order, items, nil
}
