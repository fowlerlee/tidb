package procurement

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tidb-erp-wrapper/internal/db"
	"tidb-erp-wrapper/internal/models"
)

type Service struct {
	db *db.DBHandler
}

func NewService(db *db.DBHandler) *Service {
	return &Service{db: db}
}

func (s *Service) CreateSupplier(ctx context.Context, supplier *models.Supplier) error {
	query := `
		INSERT INTO suppliers (name, code, tax_id, address, contact_info, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		supplier.Name,
		supplier.Code,
		supplier.TaxID,
		supplier.Address,
		supplier.ContactInfo,
	)
	if err != nil {
		return fmt.Errorf("error creating supplier: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	supplier.ID = id
	return nil
}

func (s *Service) CreatePurchaseOrder(ctx context.Context, po *models.PurchaseOrder, items []models.PurchaseOrderItem) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert purchase order
	query := `
		INSERT INTO purchase_orders (
			order_number, supplier_id, order_date, delivery_date,
			status, total_amount, currency_code, payment_terms,
			approved_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		po.OrderNumber,
		po.SupplierID,
		po.OrderDate,
		po.DeliveryDate,
		po.Status,
		po.TotalAmount,
		po.CurrencyCode,
		po.PaymentTerms,
		po.ApprovedBy,
	)
	if err != nil {
		return fmt.Errorf("error creating purchase order: %v", err)
	}

	poID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting purchase order ID: %v", err)
	}
	po.ID = poID

	// Insert purchase order items
	for _, item := range items {
		query := `
			INSERT INTO purchase_order_items (
				purchase_order_id, product_code, description,
				quantity, unit_price, total_price, tax_rate,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			poID,
			item.ProductCode,
			item.Description,
			item.Quantity,
			item.UnitPrice,
			item.TotalPrice,
			item.TaxRate,
		)
		if err != nil {
			return fmt.Errorf("error creating purchase order item: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Service) GetPurchaseOrderByID(ctx context.Context, id int64) (*models.PurchaseOrder, []models.PurchaseOrderItem, error) {
	po := &models.PurchaseOrder{}
	query := `
		SELECT id, order_number, supplier_id, order_date, delivery_date,
			   status, total_amount, currency_code, payment_terms,
			   approved_by, created_at, updated_at
		FROM purchase_orders
		WHERE id = ?
	`
	err := s.db.DB().QueryRowContext(ctx, query, id).Scan(
		&po.ID,
		&po.OrderNumber,
		&po.SupplierID,
		&po.OrderDate,
		&po.DeliveryDate,
		&po.Status,
		&po.TotalAmount,
		&po.CurrencyCode,
		&po.PaymentTerms,
		&po.ApprovedBy,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("purchase order not found")
		}
		return nil, nil, fmt.Errorf("error querying purchase order: %v", err)
	}

	// Get purchase order items
	items := []models.PurchaseOrderItem{}
	query = `
		SELECT id, product_code, description, quantity,
			   unit_price, total_price, tax_rate,
			   created_at, updated_at
		FROM purchase_order_items
		WHERE purchase_order_id = ?
	`
	rows, err := s.db.DB().QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying purchase order items: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.PurchaseOrderItem
		err := rows.Scan(
			&item.ID,
			&item.ProductCode,
			&item.Description,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
			&item.TaxRate,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning purchase order item: %v", err)
		}
		item.PurchaseOrderID = id
		items = append(items, item)
	}

	return po, items, nil
}
