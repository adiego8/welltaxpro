package store

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/logger"
)

// tenantConnection holds a database connection and its last access time
type tenantConnection struct {
	db         *sql.DB
	lastAccess time.Time
}

// Store manages WellTaxPro's own database and tenant connections
type Store struct {
	ctx              context.Context
	DB               *sql.DB // WellTaxPro's own database
	tenantConns      map[string]*tenantConnection
	tenantConnsMutex sync.RWMutex
	stopEviction     chan struct{}
}

// NewStore creates a new Store instance and starts the connection eviction goroutine
func NewStore(ctx context.Context, db *sql.DB) *Store {
	s := &Store{
		ctx:          ctx,
		DB:           db,
		tenantConns:  make(map[string]*tenantConnection),
		stopEviction: make(chan struct{}),
	}

	// Start background goroutine to evict idle connections
	go s.evictIdleConnections()

	return s
}

// Close closes all tenant database connections and the main database connection
func (s *Store) Close() error {
	// Stop eviction goroutine
	close(s.stopEviction)

	s.tenantConnsMutex.Lock()
	defer s.tenantConnsMutex.Unlock()

	// Close all tenant connections
	for tenantID, conn := range s.tenantConns {
		if err := conn.db.Close(); err != nil {
			logger.Errorf("Error closing connection for tenant %s: %v", tenantID, err)
		}
		delete(s.tenantConns, tenantID)
	}

	// Close main database
	return s.DB.Close()
}

// evictIdleConnections runs in background and closes connections idle for > 5 minutes
func (s *Store) evictIdleConnections() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	idleTimeout := 5 * time.Minute

	for {
		select {
		case <-s.stopEviction:
			return
		case <-ticker.C:
			s.tenantConnsMutex.Lock()
			now := time.Now()

			for tenantID, conn := range s.tenantConns {
				if now.Sub(conn.lastAccess) > idleTimeout {
					logger.Infof("Evicting idle connection for tenant %s (idle for %v)", tenantID, now.Sub(conn.lastAccess))
					if err := conn.db.Close(); err != nil {
						logger.Errorf("Error closing idle connection for tenant %s: %v", tenantID, err)
					}
					delete(s.tenantConns, tenantID)
				}
			}

			s.tenantConnsMutex.Unlock()
		}
	}
}
