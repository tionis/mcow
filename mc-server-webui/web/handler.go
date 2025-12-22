package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mc-server-webui/auth"
	"mc-server-webui/config"
	"mc-server-webui/database"
	"mc-server-webui/modmanager"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

//go:embed templates/*.html assets/*
var templateFS embed.FS

// WebHandler handles frontend requests.
type WebHandler struct {
	Store  *database.Store
	Config *config.Config
	Auth   *auth.Authenticator
}

// NewWebHandler creates a new WebHandler.
func NewWebHandler(store *database.Store, cfg *config.Config, auth *auth.Authenticator) *WebHandler {
	return &WebHandler{Store: store, Config: cfg, Auth: auth}
}

// ... (Home, ServerDetail, Admin handlers remain unchanged) ...

// FileManager renders the file manager for a specific server.
func (h *WebHandler) FileManager(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverName := vars["serverName"]

	server, err := h.Store.GetServerByName(serverName)
	if err != nil || server == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	modTree, err := modmanager.ScanModDirectory(h.Config.ModDataPath, serverName)
	// If dir not found, maybe just empty tree or create it?
	// Create if not exists to allow uploading
	if err != nil && strings.Contains(err.Error(), "not found") {
		os.MkdirAll(filepath.Join(h.Config.ModDataPath, serverName), 0755)
		modTree = &modmanager.ModItem{Name: serverName, Type: modmanager.TypeDir, Path: ""}
	} else if err != nil {
		http.Error(w, "Error scanning files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Server        *database.Server
		Authenticated bool
		UserEmail     string
		Files         *modmanager.ModItem
	}{
		Server:        server,
		Authenticated: true,
		UserEmail:     h.Auth.GetUserEmail(r),
		Files:         modTree,
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFS(templateFS, "templates/base.html", "templates/filemanager.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleFileUpload handles uploading files.
func (h *WebHandler) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	// Limit 1GB (adjust as needed)
	r.ParseMultipartForm(1024 << 20)

	serverName := r.FormValue("serverName")
	relPath := r.FormValue("path")
	
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !isValidPath(serverName, relPath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetDir := filepath.Join(h.Config.ModDataPath, serverName, relPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	targetPath := filepath.Join(targetDir, header.Filename)
	out, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/files/"+serverName, http.StatusFound)
}

// HandleFileDelete handles deleting files or directories.
func (h *WebHandler) HandleFileDelete(w http.ResponseWriter, r *http.Request) {
	serverName := r.FormValue("serverName")
	relPath := r.FormValue("path") // full relative path including filename

	if !isValidPath(serverName, relPath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetPath := filepath.Join(h.Config.ModDataPath, serverName, relPath)
	if err := os.RemoveAll(targetPath); err != nil {
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/files/"+serverName, http.StatusFound)
}

// HandleMkdir handles creating directories.
func (h *WebHandler) HandleMkdir(w http.ResponseWriter, r *http.Request) {
	serverName := r.FormValue("serverName")
	relPath := r.FormValue("path") // parent dir
	dirName := r.FormValue("dirname")

	if !isValidPath(serverName, relPath) || !isValidPath(serverName, dirName) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	targetPath := filepath.Join(h.Config.ModDataPath, serverName, relPath, dirName)
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/files/"+serverName, http.StatusFound)
}

// ServeAssets serves static assets embedded in the binary.
func (h *WebHandler) ServeAssets(w http.ResponseWriter, r *http.Request) {
	// The embed FS root contains "assets" directory.
	// We want /assets/background.jpg to map to assets/background.jpg
	// So we serve the root of templateFS.
	http.FileServer(http.FS(templateFS)).ServeHTTP(w, r)
}

func isValidPath(serverName, path string) bool {
	// prevent .. traversal
	if strings.Contains(path, "..") || strings.Contains(serverName, "..") {
		return false
	}
	// serverName restricted chars check
	for _, r := range serverName {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// ... (existing Create/Update/Delete handlers) ...

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
		if s.State != "offline" || isAuthenticated {
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
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
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

	// Access control: if offline and not admin, return 404
	if server.State == "offline" && !isAuthenticated {
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
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
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
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
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
		State:       r.FormValue("state"),
		ShowMOTD:    r.FormValue("show_motd") == "on",
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
		State:       r.FormValue("state"),
		ShowMOTD:    r.FormValue("show_motd") == "on",
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