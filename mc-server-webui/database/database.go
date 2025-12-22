package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Server represents a single Minecraft server configuration.
type Server struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
	// URL to the bluemap instance
	BlueMapURL string `json:"blueMapUrl"`
	// Path to the modpack file
	ModpackURL string `json:"modpackUrl"`
	IsEnabled  bool   `json:"isEnabled"`
}

// Store holds the database connection.
type Store struct {
	DB *sql.DB
}

// NewStore initializes the database connection and returns a Store instance.
func NewStore(dataSourceName string) (*Store, error) {
	// Run migrations before opening the main connection pool
	if err := runMigrations(dataSourceName); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection established.")
	store := &Store{DB: db}
	
	return store, nil
}

// runMigrations applies database migrations.
func runMigrations(dataSourceName string) error {
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// We need a separate connection for migration because migrate needs to see the driver
	// But passing the dsn directly is easier for sqlite3 with the migrate library
	// The migrate library URL format for sqlite3 is "sqlite3://path/to/file"
	
	// However, we are using the database/sqlite3 driver helper from migrate.
	// It requires an *sql.DB instance.
	
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()

	dbDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs", driver,
		"sqlite3", dbDriver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations ran successfully.")
	return nil
}

// ListServers retrieves all servers from the database.
func (s *Store) ListServers() ([]Server, error) {
	rows, err := s.DB.Query("SELECT id, name, address, description, blue_map_url, modpack_url, is_enabled FROM servers ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var srv Server
		var isEnabled int
		if err := rows.Scan(&srv.ID, &srv.Name, &srv.Address, &srv.Description, &srv.BlueMapURL, &srv.ModpackURL, &isEnabled); err != nil {
			return nil, err
		}
		srv.IsEnabled = isEnabled == 1
		servers = append(servers, srv)
	}

	return servers, nil
}

// GetServerByName retrieves a single server from the database by its name.
func (s *Store) GetServerByName(name string) (*Server, error) {
	row := s.DB.QueryRow("SELECT id, name, address, description, blue_map_url, modpack_url, is_enabled FROM servers WHERE name = ?", name)

	var srv Server
	var isEnabled int
	err := row.Scan(&srv.ID, &srv.Name, &srv.Address, &srv.Description, &srv.BlueMapURL, &srv.ModpackURL, &isEnabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Server not found
		}
		return nil, err
	}
	srv.IsEnabled = isEnabled == 1
	return &srv, nil
}