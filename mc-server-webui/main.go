package main

import (
	"log"
	"mc-server-webui/api"
	"mc-server-webui/config"
	"mc-server-webui/database"
	"mc-server-webui/mcstatus"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database
	store, err := database.NewStore(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("could not initialize database: %s\n", err)
	}

	// 3. Initialize Cache
	cache := mcstatus.NewServerStatusCache()

	// 4. Initialize API Handler with dependencies
	serverHandler := api.NewServerHandler(store, cfg, cache)

	router := mux.NewRouter()

	// Public API routes
	router.HandleFunc("/api/servers", serverHandler.GetServers).Methods("GET")
	router.HandleFunc("/api/servers/{serverName}/status", serverHandler.GetServerStatus).Methods("GET")
	router.HandleFunc("/api/servers/{serverName}/mods", serverHandler.GetServerMods).Methods("GET")
	router.PathPrefix("/{serverName}/map/").HandlerFunc(serverHandler.BlueMapProxy)     // BlueMap Proxy route
	router.PathPrefix("/files/{serverName}/mods/").Handler(http.HandlerFunc(serverHandler.ServeModFiles)) // Serve static mod files

	log.Printf("Starting server on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
