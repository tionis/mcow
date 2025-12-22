package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"mc-server-webui/auth"
	"mc-server-webui/database"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

//go:embed templates/*.html
var templateFS embed.FS

// WebHandler handles frontend requests.
type WebHandler struct {
	Store *database.Store
	Auth  *auth.Authenticator
}

// NewWebHandler creates a new WebHandler.
func NewWebHandler(store *database.Store, auth *auth.Authenticator) *WebHandler {
	return &WebHandler{Store: store, Auth: auth}
}

// Home renders the main server list page.
func (h *WebHandler) Home(w http.ResponseWriter, r *http.Request) {
	allServers, err := h.Store.ListServers()
	if err != nil {
		http.Error(w, "Failed to load servers", http.StatusInternalServerError)
		return
	}

	isAuthenticated := h.Auth != nil && h.Auth.IsAuthenticated(r)
	
	// Filter servers
	var visibleServers []database.Server
	for _, s := range allServers {
		if s.IsEnabled {
			visibleServers = append(visibleServers, s)
		}
	}

	data := struct {
		Servers       []database.Server
		Authenticated bool
	}{
		Servers:       visibleServers,
		Authenticated: isAuthenticated,
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"firstLine": func(s string) string {
			if idx := strings.Index(s, "\n"); idx != -1 {
				return s[:idx]
			}
			return s
		},
	}

	// Parse both base and index templates
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFS(templateFS, "templates/base.html", "templates/index.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ServerDetail renders the detail page for a specific server.
func (h *WebHandler) ServerDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["serverName"]

	server, err := h.Store.GetServerByName(serverName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if server == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	isAuthenticated := h.Auth != nil && h.Auth.IsAuthenticated(r)

	// Access control: if disabled and not admin, return 404
	if !server.IsEnabled && !isAuthenticated {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	data := struct {
		Server        *database.Server
		Authenticated bool
	}{
		Server:        server,
		Authenticated: isAuthenticated,
	}

	// Helper functions needed for base.html if it uses them, but usually only page-specific
	// Adding json and firstLine to be safe
	funcMap := template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"firstLine": func(s string) string {
			if idx := strings.Index(s, "\n"); idx != -1 {
				return s[:idx]
			}
			return s
		},
	}

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFS(templateFS, "templates/base.html", "templates/server.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Admin renders the admin dashboard.
func (h *WebHandler) Admin(w http.ResponseWriter, r *http.Request) {
	servers, err := h.Store.ListServers()
	if err != nil {
		http.Error(w, "Failed to load servers", http.StatusInternalServerError)
		return
	}

	data := struct {
		Servers       []database.Server
		Authenticated bool
		UserEmail     string
	}{
		Servers:       servers,
		Authenticated: true, // Admin page is protected, so always true
		UserEmail:     h.Auth.GetUserEmail(r),
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"firstLine": func(s string) string {
			if idx := strings.Index(s, "\n"); idx != -1 {
				return s[:idx]
			}
			return s
		},
	}

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFS(templateFS, "templates/base.html", "templates/admin.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleServerCreate handles the creation of a new server.
func (h *WebHandler) HandleServerCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	server := &database.Server{
		Name:        r.FormValue("name"),
		Address:     r.FormValue("address"),
		Description: r.FormValue("description"),
		BlueMapURL:  r.FormValue("blue_map_url"),
		ModpackURL:  r.FormValue("modpack_url"),
		IsEnabled:   r.FormValue("is_enabled") == "on",
		Metadata:    h.parseMetadata(r),
	}

	if err := h.Store.CreateServer(server); err != nil {
		http.Error(w, "Failed to create server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// HandleServerUpdate handles updating an existing server.
func (h *WebHandler) HandleServerUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	server := &database.Server{
		ID:          id,
		Name:        r.FormValue("name"),
		Address:     r.FormValue("address"),
		Description: r.FormValue("description"),
		BlueMapURL:  r.FormValue("blue_map_url"),
		ModpackURL:  r.FormValue("modpack_url"),
		IsEnabled:   r.FormValue("is_enabled") == "on",
		Metadata:    h.parseMetadata(r),
	}

	if err := h.Store.UpdateServer(server); err != nil {
		http.Error(w, "Failed to update server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// parseMetadata helper to extract metadata from form
func (h *WebHandler) parseMetadata(r *http.Request) map[string]string {
	meta := make(map[string]string)
	keys := r.Form["meta_key"]
	values := r.Form["meta_value"]

	// Ensure same length
	count := len(keys)
	if len(values) < count {
		count = len(values)
	}

	for i := 0; i < count; i++ {
		k := keys[i]
		v := values[i]
		if k != "" {
			meta[k] = v
		}
	}
	return meta
}

// HandleServerDelete handles deleting a server.
func (h *WebHandler) HandleServerDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteServer(id); err != nil {
		http.Error(w, "Failed to delete server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}