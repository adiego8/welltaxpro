package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"embed"

	"github.com/google/logger"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

//go:embed versions
var versions embed.FS

type Configuration struct {
	Database struct {
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		User       string `yaml:"user"`
		Password   string `yaml:"password"`
		InitDBName string `yaml:"initDbName"`
		DBName     string `yaml:"dbname"`
		SslMode    string `yaml:"sslmode"`
	} `yaml:"database"`
}

type migration struct {
	id          string
	description string
	sha256      string
	contents    string
}

func main() {
	fmt.Printf("Started WellTaxPro Provisioner\n")
	logger.Init("WellTaxPro", true, false, io.Discard)

	configFile := flag.String("config", "", "config file")
	flag.Parse()

	if *configFile == "" {
		logger.Errorf("--config argument is missing")
		os.Exit(1)
	}

	config, err := os.ReadFile(*configFile)
	if err != nil {
		logger.Errorf("Failed to read config file, error: %v", err)
		os.Exit(1)
	}

	var configuration Configuration
	err = yaml.Unmarshal(config, &configuration)
	if err != nil {
		logger.Errorf("Failed to unmarshal config file, error: %v", err)
		os.Exit(1)
	}

	host := configuration.Database.Host
	port := configuration.Database.Port
	user := configuration.Database.User
	password := configuration.Database.Password
	initDbName := configuration.Database.InitDBName
	dbname := configuration.Database.DBName
	sslmode := configuration.Database.SslMode

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT)
	defer cancel()

	defaultConnection := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, initDbName, sslmode)

	db, err := connectDB(defaultConnection)
	if err != nil {
		logger.Errorf("Failed to connect to database, error: %v", err)
		os.Exit(1)
	}

	logger.Infof("Checking if database %s exists", dbname)

	// Check if database exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRowContext(ctx, checkQuery, dbname).Scan(&exists)
	if err != nil {
		logger.Errorf("Failed to check if database exists, error: %v", err)
		os.Exit(1)
	}

	if !exists {
		logger.Infof("Creating database %s", dbname)
		query := fmt.Sprintf("CREATE DATABASE %s", dbname)
		_, err = db.ExecContext(ctx, query)
		if err != nil {
			logger.Errorf("Failed to create database, error: %v", err)
			os.Exit(1)
		}
	} else {
		logger.Infof("Database %s already exists, skipping creation", dbname)
	}
	db.Close()

	migrations := initialize(ctx)

	mainConnection := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

	db, err = connectDB(mainConnection)
	if err != nil {
		logger.Errorf("Failed to connect to database, error: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create migration tracking table if it doesn't exist
	err = createMigrationTable(ctx, db)
	if err != nil {
		logger.Errorf("Failed to create migration table, error: %v", err)
		os.Exit(1)
	}

	// Get already applied migrations
	appliedMigrations, err := getAppliedMigrations(ctx, db)
	if err != nil {
		logger.Errorf("Failed to get applied migrations, error: %v", err)
		os.Exit(1)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	})
	if err != nil {
		logger.Errorf("Failed to start transaction, error: %v", err)
		os.Exit(1)
	}

	for _, migration := range migrations {
		// Skip if migration already applied
		if appliedMigrations[migration.id] {
			logger.Infof("Skipping already applied migration: %s", migration.description)
			continue
		}

		logger.Infof("Applying migration: %s", migration.description)

		_, err := tx.ExecContext(ctx, migration.contents)
		if err != nil {
			logger.Errorf("Failed to apply migration, error: %v", err)
			tx.Rollback()
			os.Exit(1)
		}

		// Record migration as applied
		_, err = tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (id, description, checksum, applied_at) VALUES ($1, $2, $3, NOW())",
			migration.id, migration.description, migration.sha256)
		if err != nil {
			logger.Errorf("Failed to record migration, error: %v", err)
			tx.Rollback()
			os.Exit(1)
		}
	}

	tx.Commit()
	logger.Info("All migrations applied successfully")
}

func initialize(ctx context.Context) []migration {
	var migrations []migration

	ids := map[string]bool{}
	err := fs.WalkDir(versions, "versions", func(p string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		currentFileName := path.Base(p)
		v := strings.HasSuffix(currentFileName, ".sql") && strings.HasPrefix(currentFileName, "V")
		if !v {
			logger.Errorf("Invalid migration file name: %s", currentFileName)
			return nil
		}

		if ids[currentFileName] {
			logger.Errorf("Duplicate migration id: %s", currentFileName)
			return nil
		}
		ids[currentFileName] = true

		description := strings.TrimSuffix(strings.ReplaceAll(currentFileName, "_", " "), ".sql")

		hash := sha256.Sum256([]byte(currentFileName))

		data, err := versions.ReadFile(p)
		if err != nil {
			logger.Errorf("Failed to read file, error: %v", err)
			return nil
		}

		migrations = append(migrations, migration{
			id:          currentFileName,
			description: description,
			sha256:      fmt.Sprintf("%x", hash),
			contents:    string(data),
		})

		return nil
	})
	if err != nil {
		logger.Errorf("Failed to read migrations, error: %v", err)
		os.Exit(1)
	}

	return migrations
}

func connectDB(connection string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	logger.Info("Successfully Connected!")
	return db, nil
}

// createMigrationTable creates the schema_migrations table to track applied migrations
func createMigrationTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id VARCHAR(255) PRIMARY KEY,
			description TEXT NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// getAppliedMigrations returns a map of migration IDs that have already been applied
func getAppliedMigrations(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := db.QueryContext(ctx, "SELECT id FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		applied[id] = true
	}

	return applied, rows.Err()
}
