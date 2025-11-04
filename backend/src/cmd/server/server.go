package server

import (
	webapi "welltaxpro/src/api/web"
	"welltaxpro/src/internal/auth"
	"welltaxpro/src/internal/crypto"
	"welltaxpro/src/internal/notification"
	"welltaxpro/src/internal/store"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/logger"
	_ "github.com/lib/pq"
)

func Run(ctx context.Context) {
	args, err := parseArguments()
	if err != nil {
		logger.Fatalf("Failed parsing arguments: %v", err)
	}

	config, err := getConfiguration(args)
	if err != nil {
		logger.Fatalf("Failed getting configuration: %v", err)
	}

	// Initialize encryption system
	if err := crypto.InitEncryption(); err != nil {
		logger.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Connect to WellTaxPro database
	dbConnection := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s binary_parameters=yes",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SslMode,
	)

	logger.Info("Connecting to WellTaxPro database")

	db, err := sql.Open("postgres", dbConnection)
	if err != nil {
		logger.Fatalf("Failed connecting to database: %v", err)
	}
	defer db.Close()

	// Set up database connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		logger.Fatalf("Failed to ping database: %v", err)
	}

	logger.Info("Successfully connected to WellTaxPro database")

	// Initialize store
	store := store.NewStore(ctx, db)
	defer store.Close()

	// Initialize Firebase Auth
	logger.Info("Initializing Firebase authentication")
	authClient, err := auth.InitAuth(config.Firebase.APIKey, config.Firebase.ServiceAccountPath)
	if err != nil {
		logger.Fatalf("Failed to initialize Firebase auth: %v", err)
	}

	// Initialize Email Service
	logger.Info("Initializing email service")
	emailService := notification.NewEmailService(
		config.SendGrid.APIKey,
		config.SendGrid.DefaultFromEmail,
		config.SendGrid.DefaultFromName,
	)

	// Initialize API
	logger.Info("Starting API")
	api := webapi.NewAPI(ctx, store, authClient, emailService)
	api.InitRoutes()

	// Setup HTTP server with graceful shutdown
	addr := fmt.Sprintf(":%d", config.Server.Port)
	srv := &http.Server{
		Addr: addr,
		Handler: api.CORSHandler(webapi.CORSConfig{
			AllowedOrigins:   config.Cors.AllowedOrigins,
			AllowedMethods:   config.Cors.AllowedMethods,
			AllowedHeaders:   config.Cors.AllowedHeaders,
			AllowCredentials: config.Cors.AllowCredentials,
		}),
	}

	// Run the server in a separate goroutine
	go func() {
		logger.Infof("Server ready to accept connections on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shutdown the server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	// Context for shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
