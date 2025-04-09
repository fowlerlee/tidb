package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/config"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/db"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/services/bookkeeping"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/services/hr"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/services/materials"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/services/procurement"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/services/quality"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/services/sales"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	dbHandler, err := db.NewDBHandler(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbHandler.Close()

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbHandler.DB().PingContext(ctx); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	log.Println("Successfully connected to TiDB database")

	// Initialize services
	procurementSvc := procurement.NewService(dbHandler)
	bookkeepingSvc := bookkeeping.NewService(dbHandler)
	salesSvc := sales.NewService(dbHandler)
	materialsSvc := materials.NewService(dbHandler)
	hrSvc := hr.NewService(dbHandler)
	qualitySvc := quality.NewService(dbHandler)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Procurement endpoints
	r.POST("/suppliers", func(c *gin.Context) {
		var supplier models.Supplier
		if err := c.BindJSON(&supplier); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := procurementSvc.CreateSupplier(c.Request.Context(), &supplier); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, supplier)
	})

	r.POST("/purchase-orders", func(c *gin.Context) {
		var request struct {
			PurchaseOrder models.PurchaseOrder       `json:"purchase_order"`
			Items         []models.PurchaseOrderItem `json:"items"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := procurementSvc.CreatePurchaseOrder(c.Request.Context(), &request.PurchaseOrder, request.Items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	r.GET("/purchase-orders/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}

		po, items, err := procurementSvc.GetPurchaseOrderByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"purchase_order": po,
			"items":          items,
		})
	})

	// Bookkeeping endpoints
	r.POST("/accounts", func(c *gin.Context) {
		var account models.Account
		if err := c.BindJSON(&account); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := bookkeepingSvc.CreateAccount(c.Request.Context(), &account); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, account)
	})

	r.POST("/journal-entries", func(c *gin.Context) {
		var request struct {
			JournalEntry models.JournalEntry  `json:"journal_entry"`
			Lines        []models.JournalLine `json:"lines"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := bookkeepingSvc.CreateJournalEntry(c.Request.Context(), &request.JournalEntry, request.Lines); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	r.GET("/journal-entries/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}

		entry, lines, err := bookkeepingSvc.GetJournalEntry(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"journal_entry": entry,
			"lines":         lines,
		})
	})

	r.GET("/accounts/:id/balance", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}

		balance, err := bookkeepingSvc.GetAccountBalance(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balance": balance})
	})

	// Sales endpoints
	r.POST("/customers", func(c *gin.Context) {
		var customer models.Customer
		if err := c.BindJSON(&customer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := salesSvc.CreateCustomer(c.Request.Context(), &customer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, customer)
	})

	r.POST("/price-lists", func(c *gin.Context) {
		var request struct {
			PriceList models.PriceList       `json:"price_list"`
			Items     []models.PriceListItem `json:"items"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := salesSvc.CreatePriceList(c.Request.Context(), &request.PriceList, request.Items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	r.POST("/sales-orders", func(c *gin.Context) {
		var request struct {
			SalesOrder models.SalesOrder       `json:"sales_order"`
			Items      []models.SalesOrderItem `json:"items"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := salesSvc.CreateSalesOrder(c.Request.Context(), &request.SalesOrder, request.Items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	// Materials Management endpoints
	r.POST("/warehouses", func(c *gin.Context) {
		var warehouse models.Warehouse
		if err := c.BindJSON(&warehouse); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := materialsSvc.CreateWarehouse(c.Request.Context(), &warehouse); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, warehouse)
	})

	r.POST("/products", func(c *gin.Context) {
		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := materialsSvc.CreateProduct(c.Request.Context(), &product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, product)
	})

	r.GET("/products/low-stock", func(c *gin.Context) {
		products, err := materialsSvc.GetLowStockProducts(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	})

	// Human Resources endpoints
	r.POST("/departments", func(c *gin.Context) {
		var department models.Department
		if err := c.BindJSON(&department); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := hrSvc.CreateDepartment(c.Request.Context(), &department); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, department)
	})

	r.POST("/employees", func(c *gin.Context) {
		var employee models.Employee
		if err := c.BindJSON(&employee); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := hrSvc.CreateEmployee(c.Request.Context(), &employee); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, employee)
	})

	r.POST("/leave-requests", func(c *gin.Context) {
		var request models.LeaveRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := hrSvc.RequestLeave(c.Request.Context(), &request); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	// Quality Management endpoints
	r.POST("/quality-parameters", func(c *gin.Context) {
		var parameter models.QualityParameter
		if err := c.BindJSON(&parameter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := qualitySvc.CreateQualityParameter(c.Request.Context(), &parameter); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, parameter)
	})

	r.POST("/inspection-plans", func(c *gin.Context) {
		var request struct {
			Plan       models.InspectionPlan            `json:"plan"`
			Parameters []models.InspectionPlanParameter `json:"parameters"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := qualitySvc.CreateInspectionPlan(c.Request.Context(), &request.Plan, request.Parameters); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	r.POST("/quality-inspections", func(c *gin.Context) {
		var request struct {
			Inspection models.QualityInspection  `json:"inspection"`
			Results    []models.InspectionResult `json:"results"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := qualitySvc.CreateQualityInspection(c.Request.Context(), &request.Inspection, request.Results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, request)
	})

	// Server configuration
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}
