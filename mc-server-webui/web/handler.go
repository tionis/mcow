package web

import (
	"embed"
	"html/template"
	"mc-server-webui/auth"
	"mc-server-webui/database"
	"net/http"
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
	servers, err := h.Store.ListServers()
	if err != nil {
		http.Error(w, "Failed to load servers", http.StatusInternalServerError)
		return
	}

	data := struct {
		Servers       []database.Server
		Authenticated bool
	}{
		Servers:       servers,
		Authenticated: h.Auth != nil && h.Auth.IsAuthenticated(r),
	}

	// Parse both base and index templates
	tmpl, err := template.ParseFS(templateFS, "templates/base.html", "templates/index.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the "base.html" template (which includes "content" from index)
	// We need to ensure base.html is the entry point or layout.
	// Actually standard Go template practice is to Execute the layout.
	// My base.html defines "content" block as empty, and index.html defines it.
	// So I should Execute "base.html" (or whatever name the file has, typically the filename).
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
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

	tmpl, err := template.ParseFS(templateFS, "templates/base.html", "templates/admin.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
