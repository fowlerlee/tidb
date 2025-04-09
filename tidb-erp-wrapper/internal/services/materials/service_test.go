package materials

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialsService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)
	ctx := context.Background()

	t.Run("CreateProduct", func(t *testing.T) {
		product := &models.Product{
			Code:          "PROD001",
			Name:          "Test Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}

		err = svc.CreateProduct(ctx, product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)

		// Test creating duplicate product code
		dupProduct := &models.Product{
			Code:          "PROD001",
			Name:          "Duplicate Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, dupProduct)
		assert.Error(t, err)
	})

	t.Run("CreateWarehouse", func(t *testing.T) {
		warehouse := &models.Warehouse{
			Code:    "WH001",
			Name:    "Test Warehouse",
			Address: "123 Test St",
			Status:  "active",
		}

		err = svc.CreateWarehouse(ctx, warehouse)
		assert.NoError(t, err)
		assert.NotZero(t, warehouse.ID)

		// Test creating duplicate warehouse code
		dupWarehouse := &models.Warehouse{
			Code:    "WH001",
			Name:    "Duplicate Warehouse",
			Address: "456 Test St",
			Status:  "active",
		}
		err = svc.CreateWarehouse(ctx, dupWarehouse)
		assert.Error(t, err)
	})

	t.Run("CreateInventoryTransaction", func(t *testing.T) {
		// Create test warehouse and storage location
		warehouse := &models.Warehouse{
			Code:    "WH002",
			Name:    "Transaction Test Warehouse",
			Address: "123 Test St",
			Status:  "active",
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		location := &models.StorageLocation{
			WarehouseID: warehouse.ID,
			Code:        "LOC001",
			Name:        "Test Location",
			StorageType: "shelf",
			Capacity:    1000.0,
			Status:      "active",
		}
		err = svc.CreateStorageLocation(ctx, location)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD003",
			Name:          "Transaction Test Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Test creating a valid inventory transaction
		transaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "receipt",
			ReferenceType:     "purchase_order",
			ReferenceID:       1,
			Quantity:          100.0,
			UnitCost:          90.0,
			TransactionDate:   time.Now(),
		}

		err = svc.CreateInventoryTransaction(ctx, transaction)
		assert.NoError(t, err)
		assert.NotZero(t, transaction.ID)

		// Test issuing more than available stock
		invalidTransaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "issue",
			ReferenceType:     "sales_order",
			ReferenceID:       1,
			Quantity:          200.0, // More than available stock
			UnitCost:          90.0,
			TransactionDate:   time.Now(),
		}

		err = svc.CreateInventoryTransaction(ctx, invalidTransaction)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})

	t.Run("CreateStockMovement", func(t *testing.T) {
		// Create test warehouse
		warehouse := &models.Warehouse{
			Code:        "WH001",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		// Create test product
		category := &models.ProductCategory{
			Code:        "CAT001",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(ctx, category)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD001",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Test creating a valid stock movement (receipt)
		movement := &models.StockMovement{
			ReferenceNo:  "RCV001",
			MovementType: "receipt",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test receipt",
		}

		items := []models.StockMovementItem{
			{
				ProductID:     product.ID,
				Quantity:      100,
				UnitCost:      90.0,
				BatchNumber:   "BATCH001",
				ExpiryDate:    time.Now().AddDate(1, 0, 0),
				SerialNumbers: "SN001-SN100",
			},
		}

		err = svc.CreateStockMovement(ctx, movement, items)
		assert.NoError(t, err)
		assert.NotZero(t, movement.ID)

		// Test creating movement with invalid warehouse
		invalidMovement := &models.StockMovement{
			ReferenceNo:  "RCV002",
			MovementType: "receipt",
			WarehouseID:  999999,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test invalid receipt",
		}
		err = svc.CreateStockMovement(ctx, invalidMovement, items)
		assert.Error(t, err)

		// Test creating issue movement with insufficient stock
		issueMovement := &models.StockMovement{
			ReferenceNo:  "ISS001",
			MovementType: "issue",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test issue",
		}

		issueItems := []models.StockMovementItem{
			{
				ProductID:   product.ID,
				Quantity:    200, // More than available stock
				BatchNumber: "BATCH001",
			},
		}

		err = svc.CreateStockMovement(ctx, issueMovement, issueItems)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Create test warehouse and product
		warehouse := &models.Warehouse{
			Code:        "WH002",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		category := &models.ProductCategory{
			Code:        "CAT002",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(ctx, category)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD002",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Try to create a stock movement with invalid items to trigger rollback
		movement := &models.StockMovement{
			ReferenceNo:  "RCV003",
			MovementType: "receipt",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test rollback",
		}

		invalidItems := []models.StockMovementItem{
			{
				ProductID:   product.ID,
				Quantity:    -100, // Invalid negative quantity
				UnitCost:    90.0,
				BatchNumber: "BATCH001",
				ExpiryDate:  time.Now().AddDate(1, 0, 0),
			},
		}

		err = svc.CreateStockMovement(ctx, movement, invalidItems)
		assert.Error(t, err)

		// Verify no stock levels were affected
		var stockLevel float64
		err = testDB.DB.DB().QueryRowContext(ctx,
			"SELECT COALESCE(SUM(quantity), 0) FROM stock_levels WHERE product_id = ? AND warehouse_id = ?",
			product.ID, warehouse.ID).Scan(&stockLevel)
		require.NoError(t, err)
		assert.Equal(t, 0.0, stockLevel)
	})

	t.Run("ConcurrentStockMovements", func(t *testing.T) {
		// Create test warehouse and product
		warehouse := &models.Warehouse{
			Code:        "WH003",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		category := &models.ProductCategory{
			Code:        "CAT003",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(ctx, category)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD003",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Create initial stock receipt
		initMovement := &models.StockMovement{
			ReferenceNo:  "RCV004",
			MovementType: "receipt",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "completed",
			Notes:        "Initial stock",
		}

		initItems := []models.StockMovementItem{
			{
				ProductID:   product.ID,
				Quantity:    1000,
				UnitCost:    90.0,
				BatchNumber: "BATCH001",
				ExpiryDate:  time.Now().AddDate(1, 0, 0),
			},
		}

		err = svc.CreateStockMovement(ctx, initMovement, initItems)
		require.NoError(t, err)

		// Create multiple issue movements concurrently
		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func(i int) {
				movement := &models.StockMovement{
					ReferenceNo:  fmt.Sprintf("ISS00%d", i+1),
					MovementType: "issue",
					WarehouseID:  warehouse.ID,
					MovementDate: time.Now(),
					Status:       "pending",
					Notes:        fmt.Sprintf("Concurrent issue %d", i+1),
				}

				items := []models.StockMovementItem{
					{
						ProductID:   product.ID,
						Quantity:    100,
						BatchNumber: "BATCH001",
					},
				}

				err := svc.CreateStockMovement(ctx, movement, items)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Verify final stock level is correct (1000 - 5*100 = 500)
		var finalStock float64
		err = testDB.DB.DB().QueryRowContext(ctx,
			"SELECT quantity FROM stock_levels WHERE product_id = ? AND warehouse_id = ?",
			product.ID, warehouse.ID).Scan(&finalStock)
		require.NoError(t, err)
		assert.Equal(t, 500.0, finalStock)
	})

	t.Run("GetProductStock", func(t *testing.T) {
		// Create test category first
		category := &models.ProductCategory{
			Code:        "CAT-STOCK",
			Name:        "Stock Category",
			Description: "Test Category for Stock",
			ParentID:    nil,
			IsActive:    true,
		}
		err := svc.CreateProductCategory(ctx, category)
		require.NoError(t, err)

		// Create test warehouse
		warehouse := &models.Warehouse{
			Code:        "WH-STOCK",
			Name:        "Stock Test Warehouse",
			Description: "Test Description",
			Address:     "123 Stock St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err = svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "STOCK-TEST",
			Name:          "Stock Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			ReorderPoint:  10,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Create an inbound transaction
		transaction := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    50.0,
			Type:        "RECEIPT",
			Reference:   "TEST-RECEIPT-001",
		}
		err = svc.CreateInventoryTransaction(ctx, transaction)
		require.NoError(t, err)

		// Create an outbound transaction
		outbound := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    -20.0,
			Type:        "ISSUE",
			Reference:   "TEST-ISSUE-001",
		}
		err = svc.CreateInventoryTransaction(ctx, outbound)
		require.NoError(t, err)

		// Get stock levels
		stock, err := svc.GetProductStock(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, 30.0, stock[warehouse.ID])
	})

	t.Run("GetLowStockProducts", func(t *testing.T) {
		product := &models.Product{
			Code:          "PROD005",
			Name:          "Low Stock Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 50.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  75.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Add stock below reorder point
		transaction := &models.InventoryTransaction{
			ProductID:       product.ID,
			WarehouseID:     warehouse.ID,
			TransactionType: "receipt",
			ReferenceType:   "purchase_order",
			ReferenceID:     1,
			Quantity:        30.0,
			UnitCost:        90.0,
			TransactionDate: time.Now(),
		}
		err = svc.CreateInventoryTransaction(ctx, transaction)
		require.NoError(t, err)

		// Check low stock products
		lowStockProducts, err := svc.GetLowStockProducts(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range lowStockProducts {
			if p.ID == product.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Low stock product should be in the results")
	})

	t.Run("InventoryTransactionValidation", func(t *testing.T) {
		product := &models.Product{
			Code:          "INV-TEST",
			Name:          "Inventory Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			ReorderPoint:  10,
			IsActive:      true,
		}
		err := svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Try to create outbound transaction with insufficient stock
		invalidOutbound := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    -50.0,
			Type:        "ISSUE",
			Reference:   "TEST-INVALID-ISSUE",
		}
		err = svc.CreateInventoryTransaction(ctx, invalidOutbound)
		assert.Error(t, err, "Should not allow outbound transaction with insufficient stock")

		// Add stock and then try valid outbound
		receipt := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    100.0,
			Type:        "RECEIPT",
			Reference:   "TEST-RECEIPT",
		}
		err = svc.CreateInventoryTransaction(ctx, receipt)
		require.NoError(t, err)

		validOutbound := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    -50.0,
			Type:        "ISSUE",
			Reference:   "TEST-VALID-ISSUE",
		}
		err = svc.CreateInventoryTransaction(ctx, validOutbound)
		assert.NoError(t, err, "Should allow outbound transaction with sufficient stock")
	})
}
