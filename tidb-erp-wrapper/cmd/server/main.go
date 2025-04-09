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

	"tidb-erp-wrapper/config"
	"tidb-erp-wrapper/internal/db"
	"tidb-erp-wrapper/internal/models"
	"tidb-erp-wrapper/internal/services/bookkeeping"
	"tidb-erp-wrapper/internal/services/procurement"

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
