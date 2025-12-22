package database

import (
	"database/sql"
	"embed"
	"encoding/json"
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
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Description string            `json:"description"`
	BlueMapURL  string            `json:"blueMapUrl"`
	ModpackURL  string            `json:"modpackUrl"`
	IsEnabled   bool              `json:"isEnabled"`
	Metadata    map[string]string `json:"metadata"`
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
	rows, err := s.DB.Query("SELECT id, name, address, description, blue_map_url, modpack_url, is_enabled, metadata FROM servers ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var srv Server
		var isEnabled int
		var metadataJSON string

		if err := rows.Scan(&srv.ID, &srv.Name, &srv.Address, &srv.Description, &srv.BlueMapURL, &srv.ModpackURL, &isEnabled, &metadataJSON); err != nil {
			return nil, err
		}
		srv.IsEnabled = isEnabled == 1

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &srv.Metadata); err != nil {
				log.Printf("Warning: failed to unmarshal metadata for server %s: %v", srv.Name, err)
				srv.Metadata = make(map[string]string)
			}
		} else {
			srv.Metadata = make(map[string]string)
		}

		servers = append(servers, srv)
	}

	return servers, nil
}

// GetServerByName retrieves a single server from the database by its name.
func (s *Store) GetServerByName(name string) (*Server, error) {
	row := s.DB.QueryRow("SELECT id, name, address, description, blue_map_url, modpack_url, is_enabled, metadata FROM servers WHERE name = ?", name)

	var srv Server
	var isEnabled int
	var metadataJSON string

	err := row.Scan(&srv.ID, &srv.Name, &srv.Address, &srv.Description, &srv.BlueMapURL, &srv.ModpackURL, &isEnabled, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Server not found
		}
		return nil, err
	}
	srv.IsEnabled = isEnabled == 1

	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &srv.Metadata); err != nil {
			log.Printf("Warning: failed to unmarshal metadata for server %s: %v", srv.Name, err)
			srv.Metadata = make(map[string]string)
		}
	} else {
		srv.Metadata = make(map[string]string)
	}
	
	return &srv, nil
}

// CreateServer inserts a new server into the database.
func (s *Store) CreateServer(srv *Server) error {
	stmt, err := s.DB.Prepare("INSERT INTO servers (name, address, description, blue_map_url, modpack_url, is_enabled, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	isEnabled := 0
	if srv.IsEnabled {
		isEnabled = 1
	}

	metadataJSON, err := json.Marshal(srv.Metadata)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(srv.Name, srv.Address, srv.Description, srv.BlueMapURL, srv.ModpackURL, isEnabled, string(metadataJSON))
	return err
}

// UpdateServer updates an existing server in the database.
func (s *Store) UpdateServer(srv *Server) error {
	stmt, err := s.DB.Prepare("UPDATE servers SET name=?, address=?, description=?, blue_map_url=?, modpack_url=?, is_enabled=?, metadata=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	isEnabled := 0
	if srv.IsEnabled {
		isEnabled = 1
	}

	metadataJSON, err := json.Marshal(srv.Metadata)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(srv.Name, srv.Address, srv.Description, srv.BlueMapURL, srv.ModpackURL, isEnabled, string(metadataJSON), srv.ID)
	return err
}

// DeleteServer deletes a server from the database by ID.
func (s *Store) DeleteServer(id int) error {
	stmt, err := s.DB.Prepare("DELETE FROM servers WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}