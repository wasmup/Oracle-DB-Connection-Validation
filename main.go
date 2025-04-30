package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

var db *sql.DB

func Setup(dataSourceName string) (err error) {
	db, err = sql.Open("oracle", dataSourceName)
	if err != nil {
		slog.Error(`oracle_open`, `Err`, err)
		return
	}

	return
}

func PingContext(ctx context.Context) (err error) {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	err = db.PingContext(ctx)
	if err != nil {
		slog.Error("ping_err", "Err", err)
		return
	}

	return
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo})))
	slog.Info(`Go`, `Version`, runtime.Version(), `OS`, runtime.GOOS, `ARCH`, runtime.GOARCH, `now`, time.Now(), `Local`, time.Local)

	connStr := os.Getenv(`MY_ORACLE_DSN`)
	err := Setup(connStr)
	if err != nil {
		slog.Error(`msg`, `Err`, err)
		return
	}

	var ctx = context.Background()
	err = PingContext(ctx)
	if err != nil {
		slog.Error(`msg`, `Err`, err)
		return
	}

	poolSize := 5
	pool, err := NewConnectionPool("oracle", connStr, poolSize)
	if err != nil {
		slog.Error(`msg`, `Err`, err)
		return
	}
	defer pool.Close()

	// Acquire connection
	conn, err := pool.AcquireConnection()
	if err != nil {
		slog.Error(`msg`, `Err`, err)
		return
	}

	// Use the connection (example query)
	ctx = context.Background()
	err = conn.PingContext(ctx)
	if err != nil {
		fmt.Printf("Ping failed: %v\n", err)
	} else {
		fmt.Println("Ping succeeded")
	}

	// Release connection back to pool
	pool.ReleaseConnection(conn)
}

// ConnectionPool manages a pool of *sql.DB connections.
type ConnectionPool struct {
	mu          sync.Mutex
	connections chan *sql.DB
	connStr     string
	driverName  string
	poolSize    int
}

// NewConnectionPool initializes the pool with open connections.
func NewConnectionPool(driverName, connStr string, poolSize int) (*ConnectionPool, error) {
	if poolSize <= 0 {
		return nil, errors.New("poolSize must be > 0")
	}

	pool := &ConnectionPool{
		connections: make(chan *sql.DB, poolSize),
		connStr:     connStr,
		driverName:  driverName,
		poolSize:    poolSize,
	}

	// Open initial connections and add to the pool
	for range poolSize {
		db, err := sql.Open(driverName, connStr)
		if err != nil {
			// Close all previously opened connections on error
			pool.Close()
			return nil, fmt.Errorf("failed to open connection: %w", err)
		}
		// Optionally set connection pool settings here
		db.SetMaxOpenConns(1) // single connection per *sql.DB here
		db.SetConnMaxLifetime(time.Hour)

		pool.connections <- db
	}

	return pool, nil
}

// AcquireConnection gets a connection from the pool and ensures it is alive.
func (p *ConnectionPool) AcquireConnection() (*sql.DB, error) {
	conn := <-p.connections

	// Ping to check if connection is alive
	if err := conn.Ping(); err != nil {
		// Connection is stale or dead, close and replace it
		conn.Close()

		newConn, err := sql.Open(p.driverName, p.connStr)
		if err != nil {
			// Return the broken connection back to the pool to avoid leak
			p.connections <- conn
			return nil, fmt.Errorf("failed to open new connection: %w", err)
		}
		newConn.SetMaxOpenConns(1)
		newConn.SetConnMaxLifetime(time.Hour)

		conn = newConn
	}

	return conn, nil
}

// ReleaseConnection returns the connection back to the pool.
func (p *ConnectionPool) ReleaseConnection(conn *sql.DB) {
	p.connections <- conn
}

// Close closes all connections in the pool.
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
}
