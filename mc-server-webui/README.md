# MC Server WebUI

A modern, high-performance web interface for managing and showcasing IEEE Passau's Minecraft servers. Built with Go, it provides a clean UI for server status, mod downloads, and map viewing, protected by OIDC authentication for administrative tasks.

## Features

*   **Real-time Server Status:** Live player counts, version info, and online status using efficient caching (60s TTL).
*   **Mod File Browser:** Automatically scans and serves mod files, modpacks, and documentation from a structured directory.
*   **BlueMap Proxy:** Securely proxies BlueMap instances (e.g., `http://localhost:8100`) through the main web server, unifying access.
*   **OIDC Authentication:** Secure login via OpenID Connect (e.g., Keycloak, Google) for administrative access.
*   **Modern Architecture:**
    *   **Backend:** Go (1.21+) with `gorilla/mux` and `database/sql`.
    *   **Database:** SQLite with `golang-migrate` for robust schema management.
    *   **Frontend:** Server-side rendered HTML (Go templates) with Bootstrap 5.
    *   **Config:** 12-factor app design using environment variables.

## Getting Started

### Prerequisites

*   **Go:** Version 1.21 or higher.
*   **SQLite:** (Optional) For inspecting the database manually.
*   **GCC:** Required for `go-sqlite3` (CGO).

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/ieee-passau/mc-server-webui.git
    cd mc-server-webui
    ```

2.  Install dependencies:
    ```bash
    go mod download
    ```

3.  Build the application:
    ```bash
    go build -o mc-webui .
    ```

### Running the Application

1.  **Set Environment Variables:**
    Create a `.env` file or export these variables:
    ```bash
    export PORT=8080
    export DB_PATH=./mc-servers.db
    export MOD_DATA_PATH=./data/mods
    # OIDC Config (Optional - Login disabled if missing)
    export OIDC_PROVIDER_URL=https://auth.example.com/realms/ieee
    export OIDC_CLIENT_ID=mc-webui
    export OIDC_CLIENT_SECRET=your-secret
    export OIDC_REDIRECT_URL=http://localhost:8080/auth/callback
    export SESSION_SECRET=change-this-to-a-long-random-string
    ```

2.  **Run the binary:**
    ```bash
    ./mc-webui
    ```
    Or directly with Go:
    ```bash
    go run .
    ```

3.  Access the UI at `http://localhost:8080`.

## Configuration Reference

| Variable             | Default                         | Description                                                                 |
| -------------------- | ------------------------------- | --------------------------------------------------------------------------- |
| `PORT`               | `8080`                          | The HTTP port to listen on.                                                 |
| `DB_PATH`            | `./mc-servers.db`               | Path to the SQLite database file. Created automatically if missing.         |
| `MOD_DATA_PATH`      | `data/mods`                     | Root directory for storing server mod files.                                |
| `OIDC_PROVIDER_URL`  | *(Empty)*                       | The OIDC Issuer URL (e.g., Keycloak realm URL). Login disabled if empty.    |
| `OIDC_CLIENT_ID`     | *(Empty)*                       | The Client ID registered with your IDP.                                     |
| `OIDC_CLIENT_SECRET` | *(Empty)*                       | The Client Secret for the application.                                      |
| `OIDC_REDIRECT_URL`  | `.../auth/callback`             | The callback URL whitelisted in your IDP.                                   |
| `SESSION_SECRET`     | `super-secret...`               | Random string used to encrypt session cookies. **Change in production!**    |

## Usage Guide

### 1. Adding Servers
Currently, servers are managed via direct database insertion (Admin UI coming soon).
To add a server manually:
```bash
sqlite3 mc-servers.db "INSERT INTO servers (name, address, description, blue_map_url, is_enabled) VALUES ('Creative', 'mc.example.com:25565', 'Our creative world', 'http://map-host:8100', 1);"
```

### 2. Organizing Mod Files
The application serves files from `MOD_DATA_PATH` (default: `data/mods`).
Create a directory matching the **exact server name** (case-sensitive) from the database:

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

*   **`.md` files:** The content is read and sent to the frontend (rendering support in UI to be added).
*   **`.url` files:** Contain a single line with a URL. The UI renders these as external links.
*   **Other files:** Served as direct downloads.

### 3. BlueMap Proxy
To enable the map proxy:
1.  Ensure your BlueMap backend is running (e.g., internal IP `10.0.0.5:8100`).
2.  Update the server record in the database:
    ```sql
    UPDATE servers SET blue_map_url = 'http://10.0.0.5:8100' WHERE name = 'Creative';
    ```
3.  The map will be accessible at `http://localhost:8080/Creative/map/`.

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

## Development

### Running Migrations
Migrations run automatically on startup. To add a new migration:
1.  Create a pair of files in `database/migrations/`:
    *   `XXXXXX_description.up.sql`
    *   `XXXXXX_description.down.sql`
2.  Rebuild the application (migrations are embedded).

### License
[MIT License](LICENSE)