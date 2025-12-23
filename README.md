# MineCraft Overview (McOw/MCOW)

A modern, high-performance web interface for managing and showcasing IEEE Passau's Minecraft servers. Built with Go, it provides a clean UI for server status, mod downloads, and map viewing, protected by OIDC authentication for administrative tasks.

## Features

*   **Real-time Server Status:** Live player counts, version info, and online status using efficient caching (60s TTL).
*   **Admin Dashboard:** Complete web-based management interface for adding, editing, and deleting servers without touching the database.
*   **Mod File Browser:** Automatically scans and serves mod files, modpacks, and documentation from a structured directory. Supports downloading files and directories (as zip), rendering `.md` files, and `.url` redirects.
*   **BlueMap Proxy:** Securely proxies BlueMap instances (e.g., `http://localhost:8100`) through the main web server, unifying access.
*   **OIDC Authentication:** Secure login via OpenID Connect (e.g., Keycloak, Google) for administrative access.
*   **Modern Architecture:**
    *   **Backend:** Go (1.24+) with `gorilla/mux` and `database/sql`.
    *   **Database:** SQLite with `golang-migrate` for robust schema management.
    *   **Frontend:** Server-side rendered HTML (Go templates) with Bootstrap 5.
    *   **Config:** 12-factor app design using environment variables.

## Getting Started

### Prerequisites

*   **Go:** Version 1.24 or higher.
*   **SQLite:** (Optional) For inspecting the database manually.
*   **GCC:** Required for `go-sqlite3` (CGO).

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/tionis/mcow.git
    cd mcow
    ```

2.  Build the application:
    ```bash
    go build -o mcow .
    ```

### Containers (Docker/Podman)

#### Pre-built Image
We automatically build a container image for the `main` branch, available at `ghcr.io/tionis/mcow:latest`.

#### Run with Podman
Here is an example of how to run the application using Podman, persisting data and mods:

```bash
podman run -d --name mcow \
    -p 8080:8080 \
    -v ./data:/data \
    -v ./mods:/app/data/mods \
    -e SESSION_SECRET="change-this-to-a-long-random-string" \
    -e OIDC_PROVIDER_URL="" \
    -e OIDC_CLIENT_ID="" \
    -e OIDC_CLIENT_SECRET="" \
    ghcr.io/tionis/mcow:latest
```
*Note: Refer to the [Configuration Reference](#configuration-reference) below for OIDC details required for admin access.*

#### Build Manually
A `Containerfile` is provided for building a container image manually:
```bash
podman build -t mcow .
podman run -p 8080:8080 -v ./data:/data mcow
```

### Running the Application

1.  **Set Environment Variables:**
    Create a `.env` file or export these variables:
    ```bash
    export PORT=8080
    export DB_PATH=./mcow.db
    export MOD_DATA_PATH=./data/mods
    # OIDC Config (Optional - Login disabled if missing)
    export OIDC_PROVIDER_URL=https://auth.example.com/realms/ieee
    export OIDC_CLIENT_ID=mcow
    export OIDC_CLIENT_SECRET=your-secret
    export OIDC_REDIRECT_URL=http://localhost:8080/auth/callback
    export SESSION_SECRET=change-this-to-a-long-random-string
    ```

2.  **Run the binary:**
    ```bash
    ./mcow
    ```
    Or directly with Go:
    ```bash
    go run .
    ```

3.  Access the UI at `http://localhost:8080`.
    *   **Admin Login:** Go to `/login` to authenticate via OIDC (if configured).
    *   **Admin Dashboard:** Go to `/admin` to manage servers.

### Helper Scripts
The repository includes helper scripts for development/testing:
*   `go run insert_dummy_data.go`: Populates the database with sample servers.
*   `go run update_bluemap_url.go`: Example script to update database records programmatically.

## Configuration Reference



| Variable             | Default                         | Description                                                                 |
| -------------------- | ------------------------------- | --------------------------------------------------------------------------- |
| `PORT`               | `8080`                          | The HTTP port to listen on.                                                 |
| `DB_PATH`            | `./mcow.db`                     | Path to the SQLite database file. Created automatically if missing.         |
| `MOD_DATA_PATH`      | `data/mods`                     | Root directory for storing server mod files.                                |
| `OIDC_PROVIDER_URL`  | *(Empty)*                       | The OIDC Issuer URL (e.g., Keycloak realm URL). Login disabled if empty.    |
| `OIDC_CLIENT_ID`     | *(Empty)*                       | The Client ID registered with your IDP.                                     |
| `OIDC_CLIENT_SECRET` | *(Empty)*                       | The Client Secret for the application.                                      |
| `OIDC_REDIRECT_URL`  | `.../auth/callback`             | The callback URL whitelisted in your IDP.                                   |
| `SESSION_SECRET`     | `super-secret...`               | Random string used to encrypt session cookies. **Change in production!**    |

## Usage Guide

### 1. Managing Servers
Servers are managed via the web-based Admin Dashboard at `/admin`.
1.  **Log in:** Authenticate via OIDC to access the dashboard.
2.  **Add/Edit:** Use the interface to configure server details:
    *   **Name:** Unique identifier (used in URLs and file paths).
    *   **Address:** The Minecraft server address (e.g., `mc.example.com`).
    *   **State:** Controls visibility (`online`, `offline`, `planned`, `maintenance`).
    *   **BlueMap URL:** Internal URL for proxying (e.g., `http://localhost:8100`).
    *   **Modpack URL:** Optional direct download link.
    *   **Metadata:** Custom key-value pairs for additional info.

### 2. Managing Files
You can manage mod files via the **Admin File Manager** or directly on the filesystem.

#### Via File Manager
Click "Files" on any server in the Admin Dashboard to upload, delete, or organize files directly from the browser.

#### Manual Organization
The application serves files from `MOD_DATA_PATH` (default: `data/mods`).
Directory structure must match the **server name**:

```text
data/mods/
├── Creative/               <-- Matches server name "Creative"
│   ├── mods/
│   │   ├── sodium.jar
│   │   └── iris.jar
│   ├── modpack-v1.zip
│   ├── rules.md            <-- Rendered as text/markdown
│   └── discord.url         <-- Rendered as a link
└── Survival/
    └── ...
```

*   **`.md` files:** Content is displayed in the file browser.
*   **`.url` files:** Rendered as external links.
*   **Other files:** Served as direct downloads.

### 3. BlueMap Proxy
To enable the map proxy:
1.  Ensure your BlueMap backend is running (e.g., internal IP `10.0.0.5:8100`).
2.  In the Admin Dashboard, set the **BlueMap URL** for the server to `http://10.0.0.5:8100`.
3.  The map will be accessible publicly at `http://your-site.com/Creative/map/`.

## Architecture

*   **`main.go`**: Entry point. Wires dependencies (Config, Store, Auth) and starts the server.
*   **`api/`**: Contains API handlers (`ServerHandler`) for JSON endpoints and proxy logic.
*   **`web/`**: Contains the `WebHandler` and embedded HTML `templates/`.
*   **`auth/`**: Handles OIDC flow and session management.
*   **`database/`**:
    *   `database.go`: Connection pooling and repository pattern implementation.
    *   `migrations/`: SQL migration files embedded into the binary.
*   **`mcstatus/`**: Logic for querying Minecraft servers and caching results.
*   **`modmanager/`**: Secure filesystem scanning for mod files.
*   **`config/`**: Environment variable loading.

## API Reference

The application exposes several JSON endpoints:

*   `GET /api/servers`: Returns a list of all visible servers.
*   `GET /api/servers/{serverName}/status`: Returns real-time status (online/offline, players) for a server.
*   `GET /api/servers/{serverName}/mods`: Returns the file tree of mods for a server.
*   `GET /files/{serverName}/mods/...`: Downloads a file directly.

## Development


### Running Migrations
Migrations run automatically on startup. To add a new migration:
1.  Create a pair of files in `database/migrations/`:
    *   `XXXXXX_description.up.sql`
    *   `XXXXXX_description.down.sql`
2.  Rebuild the application (migrations are embedded).

### Running Tests
To run the test suite:
```bash
go test ./...
```
To run tests with coverage:
```bash
go test -cover ./...
```

### License
[MIT License](LICENSE)
