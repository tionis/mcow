package main

import (
	"log"
	"mc-server-webui/api"
	"mc-server-webui/auth"
	"mc-server-webui/config"
	"mc-server-webui/database"
	"mc-server-webui/mcstatus"
	"mc-server-webui/web"
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

	// 4. Initialize Authenticator
	authenticator, err := auth.NewAuthenticator(cfg)
	if err != nil {
		log.Printf("Warning: OIDC authentication could not be initialized: %v", err)
	} else if authenticator == nil {
		log.Println("OIDC authentication is disabled (not configured).")
	} else {
		log.Println("OIDC authentication initialized.")
	}

	// 5. Initialize Handlers
	serverHandler := api.NewServerHandler(store, cfg, cache, authenticator)
	webHandler := web.NewWebHandler(store, cfg, authenticator)

	router := mux.NewRouter()

	// Web Routes
	router.HandleFunc("/", webHandler.Home).Methods("GET")
	
	if authenticator != nil {
		router.HandleFunc("/login", authenticator.HandleLogin).Methods("GET")
		router.HandleFunc("/logout", authenticator.HandleLogout).Methods("GET")
		router.HandleFunc("/auth/callback", authenticator.HandleCallback).Methods("GET")
		
		// Protected Admin Route
		router.Handle("/admin", authenticator.Middleware(http.HandlerFunc(webHandler.Admin))).Methods("GET")
		router.Handle("/admin/servers/add", authenticator.Middleware(http.HandlerFunc(webHandler.HandleServerCreate))).Methods("POST")
		router.Handle("/admin/servers/update", authenticator.Middleware(http.HandlerFunc(webHandler.HandleServerUpdate))).Methods("POST")
		router.Handle("/admin/servers/delete", authenticator.Middleware(http.HandlerFunc(webHandler.HandleServerDelete))).Methods("POST")
		
		// File Manager Routes
		router.Handle("/admin/files/{serverName}", authenticator.Middleware(http.HandlerFunc(webHandler.FileManager))).Methods("GET")
		router.Handle("/admin/files/upload", authenticator.Middleware(http.HandlerFunc(webHandler.HandleFileUpload))).Methods("POST")
		router.Handle("/admin/files/delete", authenticator.Middleware(http.HandlerFunc(webHandler.HandleFileDelete))).Methods("POST")
		router.Handle("/admin/files/mkdir", authenticator.Middleware(http.HandlerFunc(webHandler.HandleMkdir))).Methods("POST")
	}

	// API Routes
	router.HandleFunc("/api/servers", serverHandler.GetServers).Methods("GET")
	router.HandleFunc("/api/servers/{serverName}/status", serverHandler.GetServerStatus).Methods("GET")
	router.HandleFunc("/api/servers/{serverName}/mods", serverHandler.GetServerMods).Methods("GET")
	router.PathPrefix("/{serverName}/map/").HandlerFunc(serverHandler.BlueMapProxy)     // BlueMap Proxy route
	router.PathPrefix("/files/{serverName}/mods/").Handler(http.HandlerFunc(serverHandler.ServeModFiles)) // Serve static mod files
	
	// Server Detail Page (catch-all for server names)
	router.HandleFunc("/{serverName}", webHandler.ServerDetail).Methods("GET")

	log.Printf("Starting server on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
