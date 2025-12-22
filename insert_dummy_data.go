//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./mcow.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	servers := []struct {
		Name        string
		Address     string
		Description string
		BlueMapURL  string
		ModpackURL  string
	}{
		{
			Name:        "Creative",
			Address:     "creative.example.com:25565",
			Description: "A server for creative building.",
			BlueMapURL:  "https://maps.example.com/creative",
			ModpackURL:  "/files/creative/mods/creative-pack-v1.zip",
		},
		{
			Name:        "Survival",
			Address:     "survival.example.com:25565",
			Description: "A challenging survival server.",
			BlueMapURL:  "https://maps.example.com/survival",
			ModpackURL:  "/files/survival/mods/survival-pack-v2.zip",
		},
	}

	insertSQL := `INSERT INTO servers (name, address, description, blue_map_url, modpack_url, is_enabled) VALUES (?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()

	for _, server := range servers {
		_, err := statement.Exec(server.Name, server.Address, server.Description, server.BlueMapURL, server.ModpackURL, 1)
		if err != nil {
			log.Printf("Could not insert server %s: %v", server.Name, err)
		} else {
			log.Printf("Inserted server: %s", server.Name)
		}
	}
	log.Println("Dummy data insertion complete.")
}
