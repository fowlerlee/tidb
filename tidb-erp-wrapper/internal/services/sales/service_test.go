package sales

import (
	"context"
	"testing"
	"time"
	"tidb-erp-wrapper/internal/models"
	"tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSalesService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)

	t.Run("CreateCustomer", func(t *testing.T) {
		// Test creating a valid customer
		customer := testutil.GenerateCustomer(0)
		err := svc.CreateCustomer(context.Background(), customer)
		assert.NoError(t, err)
		assert.NotZero(t, customer.ID)

		// Test creating duplicate customer code
		dupCustomer := testutil.GenerateCustomer(0)
		dupCustomer.Code = customer.Code
		err = svc.CreateCustomer(context.Background(), dupCustomer)
		assert.Error(t, err)

		// Test creating customer with invalid credit limit
		invalidCustomer := testutil.GenerateCustomer(0)
		invalidCustomer.CreditLimit = -1000.0
		err = svc.CreateCustomer(context.Background(), invalidCustomer)
		assert.Error(t, err)
	})

	t.Run("CreatePriceList", func(t *testing.T) {
		// Test creating a valid price list
		priceList := &models.PriceList{
			Name:         "Test Price List",
			Description:  "Test Description",
			CurrencyCode: "USD",
			StartDate:    time.Now(),
			IsActive:     true,
		}

		items := []models.PriceListItem{
			{
				ProductCode:      "PROD1",
				UnitPrice:       100.0,
				MinQuantity:     1,
				DiscountPercent: 0,
			},
			{
				ProductCode:      "PROD2",
				UnitPrice:       200.0,
				MinQuantity:     5,
				DiscountPercent: 10,
			},
		}

		err := svc.CreatePriceList(context.Background(), priceList, items)
		assert.NoError(t, err)
		assert.NotZero(t, priceList.ID)

		// Test creating price list with no items
		emptyPriceList := &models.PriceList{
			Name:         "Empty Price List",
			Description:  "Test Description",
			CurrencyCode: "USD",
			StartDate:    time.Now(),
			IsActive:     true,
		}
		err = svc.CreatePriceList(context.Background(), emptyPriceList, nil)
		assert.Error(t, err)

		// Test creating price list with invalid prices
		invalidItems := []models.PriceListItem{
			{
				ProductCode:      "PROD1",
				UnitPrice:       -100.0, // Invalid negative price
				MinQuantity:     1,
				DiscountPercent: 0,
			},
		}
		err = svc.CreatePriceList(context.Background(), priceList, invalidItems)
		assert.Error(t, err)
	})

	t.Run("CreateSalesOrder", func(t *testing.T) {
		// Create test customer first
		customer := testutil.GenerateCustomer(0)
		err := svc.CreateCustomer(context.Background(), customer)
		require.NoError(t, err)

		// Create test price list
		priceList := &models.PriceList{
			Name:         "Test Price List",
			Description:  "Test Description",
			CurrencyCode: "USD",
			StartDate:    time.Now(),
			IsActive:     true,
		}
		priceListItems := []models.PriceListItem{
			{
				ProductCode:      "PROD1",
				UnitPrice:       100.0,
				MinQuantity:     1,
				DiscountPercent: 0,
			},
		}
		err = svc.CreatePriceList(context.Background(), priceList, priceListItems)
		require.NoError(t, err)

		// Test creating a valid sales order
		order := &models.SalesOrder{
			CustomerID:   customer.ID,
			PriceListID: priceList.ID,
			OrderDate:    time.Now(),
			Status:      "draft",
			TotalAmount: 1000.0,
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}

		items := []models.SalesOrderItem{
			{
				ProductCode:     "PROD1",
				Description:    "Test Product",
				Quantity:       10,
				UnitPrice:     100.0,
				DiscountPercent: 0,
				TotalPrice:    1000.0,
				TaxRate:       0.1,
			},
		}

		err = svc.CreateSalesOrder(context.Background(), order, items)
		assert.NoError(t, err)
		assert.NotZero(t, order.ID)

		// Test creating order exceeding customer credit limit
		exceedingOrder := &models.SalesOrder{
			CustomerID:   customer.ID,
			PriceListID: priceList.ID,
			OrderDate:    time.Now(),
			Status:      "draft",
			TotalAmount: customer.CreditLimit + 1000.0, // Exceeds credit limit
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}

		err = svc.CreateSalesOrder(context.Background(), exceedingOrder, items)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds customer credit limit")

		// Test creating order with invalid customer
		invalidOrder := &models.SalesOrder{
			CustomerID:   999999,
			PriceListID: priceList.ID,
			OrderDate:    time.Now(),
			Status:      "draft",
			TotalAmount: 1000.0,
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}

		err = svc.CreateSalesOrder(context.Background(), invalidOrder, items)
		assert.Error(t, err)

		// Test creating order with mismatched totals
		mismatchOrder := &models.SalesOrder{
			CustomerID:   customer.ID,
			PriceListID: priceList.ID,
			OrderDate:    time.Now(),
			Status:      "draft",
			TotalAmount: 2000.0, // Doesn't match item totals
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}

		err = svc.CreateSalesOrder(context.Background(), mismatchOrder, items)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total amount mismatch")
	})

	t.Run("CreateDelivery", func(t *testing.T) {
		// Create test customer
		customer := testutil.GenerateCustomer(0)
		err := svc.CreateCustomer(context.Background(), customer)
		require.NoError(t, err)

		// Create test sales order
		order := &models.SalesOrder{
			CustomerID:   customer.ID,
			OrderDate:    time.Now(),
			Status:      "confirmed", // Order must be confirmed for delivery
			TotalAmount: 1000.0,
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}

		orderItems := []models.SalesOrderItem{
			{
				ProductCode:     "PROD1",
				Description:    "Test Product",
				Quantity:       10,
				UnitPrice:     100.0,
				TotalPrice:    1000.0,
				TaxRate:       0.1,
			},
		}

		err = svc.CreateSalesOrder(context.Background(), order, orderItems)
		require.NoError(t, err)

		// Test creating a valid delivery
		delivery := &models.Delivery{
			DeliveryNumber:  "DEL001",
			SalesOrderID:    order.ID,
			DeliveryDate:    time.Now(),
			Status:         "pending",
			ShippingAddress: "123 Test St",
			Carrier:        "Test Carrier",
		}

		items := []models.DeliveryItem{
			{
				SalesOrderItemID: 1, // This should match the actual sales order item ID
				Quantity:         10,
				BatchNumber:      "BATCH001",
				SerialNumbers:    "SN001-SN010",
			},
		}

		err = svc.CreateDelivery(context.Background(), delivery, items)
		assert.NoError(t, err)
		assert.NotZero(t, delivery.ID)

		// Test creating delivery for unconfirmed order
		unconfirmedOrder := &models.SalesOrder{
			CustomerID:   customer.ID,
			OrderDate:    time.Now(),
			Status:      "draft",
			TotalAmount: 1000.0,
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}
		err = svc.CreateSalesOrder(context.Background(), unconfirmedOrder, orderItems)
		require.NoError(t, err)

		invalidDelivery := &models.Delivery{
			DeliveryNumber:  "DEL002",
			SalesOrderID:    unconfirmedOrder.ID,
			DeliveryDate:    time.Now(),
			Status:         "pending",
			ShippingAddress: "123 Test St",
			Carrier:        "Test Carrier",
		}

		err = svc.CreateDelivery(context.Background(), invalidDelivery, items)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot create delivery for unconfirmed")

		// Test creating delivery with quantity exceeding order
		excessItems := []models.DeliveryItem{
			{
				SalesOrderItemID: 1,
				Quantity:         20, // Exceeds order quantity of 10
				BatchNumber:      "BATCH001",
				SerialNumbers:    "SN001-SN020",
			},
		}

		excessDelivery := &models.Delivery{
			DeliveryNumber:  "DEL003",
			SalesOrderID:    order.ID,
			DeliveryDate:    time.Now(),
			Status:         "pending",
			ShippingAddress: "123 Test St",
			Carrier:        "Test Carrier",
		}

		err = svc.CreateDelivery(context.Background(), excessDelivery, excessItems)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds ordered quantity")
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Create test customer
		customer := testutil.GenerateCustomer(0)
		err := svc.CreateCustomer(context.Background(), customer)
		require.NoError(t, err)

		// Try to create an order with invalid items to trigger rollback
		order := &models.SalesOrder{
			CustomerID:   customer.ID,
			OrderDate:    time.Now(),
			Status:      "draft",
			TotalAmount: 1000.0,
			CurrencyCode: "USD",
			PaymentTerms: "Net 30",
		}

		invalidItems := []models.SalesOrderItem{
			{
				ProductCode:     "", // Invalid product code
				Description:    "Test Product",
				Quantity:       -1, // Invalid quantity
				UnitPrice:     100.0,
				TotalPrice:    1000.0,
				TaxRate:       0.1,
			},
		}

		err = svc.CreateSalesOrder(context.Background(), order, invalidItems)
		assert.Error(t, err)

		// Verify the order was not created
		_, _, err = svc.GetSalesOrderByID(context.Background(), order.ID)
		assert.Error(t, err)
	})

	t.Run("ConcurrentOrders", func(t *testing.T) {
		// Create test customer
		customer := testutil.GenerateCustomer(0)
		customer.CreditLimit = 10000.0 // High credit limit for concurrent testing
		err := svc.CreateCustomer(context.Background(), customer)
		require.NoError(t, err)

		// Create multiple orders concurrently
		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func(orderNum int) {
				order := &models.SalesOrder{
					CustomerID:   customer.ID,
					OrderDate:    time.Now(),
					Status:      "draft",
					TotalAmount: 1000.0,
					CurrencyCode: "USD",
					PaymentTerms: "Net 30",
				}

				items := []models.SalesOrderItem{
					{
						ProductCode:     fmt.Sprintf("PROD%d", orderNum),
						Description:    "Test Product",
						Quantity:       10,
						UnitPrice:     100.0,
						TotalPrice:    1000.0,
						TaxRate:       0.1,
					},
				}

				err := svc.CreateSalesOrder(context.Background(), order, items)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Verify credit limit was properly managed
		var totalOrderAmount float64
		rows, err := testDB.DB.DB().QueryContext(context.Background(),
			"SELECT COALESCE(SUM(total_amount), 0) FROM sales_orders WHERE customer_id = ?",
			customer.ID)
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())
		err = rows.Scan(&totalOrderAmount)
		require.NoError(t, err)
		assert.LessOrEqual(t, totalOrderAmount, customer.CreditLimit)
	})
})