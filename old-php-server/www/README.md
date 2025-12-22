# IEEE SB Passau Minecraft Server Website

This is a dynamic PHP-powered website to show players information about our Minecraft servers, including BlueMap interactive maps and downloadable mods/resources.

## Structure

```
/var/www/bluemap/
├── index.php          # Main landing page listing all servers (and router)
├── server.php         # Individual server detail page
├── config.php         # Configuration and helper functions
├── servers.yml        # Server configuration (EDIT THIS)
├── .htaccess          # Apache rewrite rules
│
├── data/              # Data directory for server files
│   └── servername/
│       └── mods/      # Server-specific files and downloads
│           ├── file.zip
│           └── ...
```

## Configuration

### servers.yml

This file defines all your Minecraft servers:

```yaml
---
servers:
  - name: "servername"
    bluemap_proxy: "10.0.0.60:23350"
    description: "Server description"
    modloader: "forge"  # Optional: forge, fabric, quilt, etc.
    minecraft_version: "1.20.1"  # Optional
    status: "Online"  # Online, Planned, or Offline
    server_address: "server.example.com"
    enable_query: true # Optional: enable full query (requires enable-query=true in server.properties)
    current_version_path: "v1.2.3"  # Optional: highlights this folder as current version
```

**Status values:**
- `Online` - Server is active (green badge)
- `Planned` - Server is in planning/development (orange badge)
- `Offline` - Server is not available (red badge)

**Optional fields:**
- `modloader` - Displayed on server cards and detail page
- `minecraft_version` - Displayed on server cards and detail page
- `current_version_path` - Path relative to mods folder to highlight as current version (e.g., "v1.2.3" or "versions/current")
- `enable_query` - Set to `true` to use the GameSpy4 protocol for full player lists and plugins (requires `enable-query=true` in `server.properties`).

### Mods Directory

Place downloadable files in `/var/www/bluemap/data/servername/mods/`:

The downloads section displays a **recursive folder tree** that you can expand/collapse:

```
data/
└── servername/
    └── mods/
        ├── v1.2.3/           # Folders are collapsible
        │   ├── modpack.zip
        │   └── mods/
        │       ├── mod1.jar
        │       └── mod2.jar
        ├── v1.2.2/
        │   └── modpack.zip
        └── README.txt
```

**Regular files** (.zip, .jar, .mrpack, etc.):
- Displayed in folder tree structure
- Shows file size
- Click to download

**Folders**:
- Click folder name to expand/collapse
- Supports unlimited nesting levels
- Folders containing `current_version_path` auto-expand and show "Current" badge

**External links** (.url files):
Create a Windows `.url` file format:
```ini
[InternetShortcut]
URL=https://modrinth.com/modpack/example
```

The filename (without .url) becomes the link text and appears in the tree with a link icon.

## URL Structure

- `/` - Main page with all servers
- `/servername/` - Server detail page
- `/servername/map/` - BlueMap interactive map (proxied)
- `/servername/mods/filename.ext` - Download files

## PHP Requirements

- PHP 7.4+ (PHP 8.x recommended)
- No external dependencies required (uses built-in YAML parser)
- Optional: Symfony YAML component for better YAML parsing

## Dependencies

The project includes three libraries in the `deps/` folder:

1. **PHP Minecraft Query** (`deps/mc_query`): Used to query Minecraft servers for status and player information.
   - [GitHub Repository](https://github.com/xPaw/PHP-Minecraft-Query)
2. **PHP Source Query** (`deps/source_query`): Used for RCON connections (Source Engine Query).
   - [GitHub Repository](https://github.com/xPaw/PHP-Source-Query)
3. **Parsedown** (`deps/parsedown`): Used to render Markdown files in the browser.
   - [GitHub Repository](https://github.com/erusev/parsedown)

## Deployment

The website is deployed via Ansible playbook:
```bash
ansible-playbook playbooks/bluemap-php.yml
```

**Note:** This deploys the PHP files to `/var/www/bluemap/` where they are served by a Pterodactyl-managed container running nginx + PHP-FPM.

### Web Server Configuration

**Requirements:**
- PHP 7.4+ with cURL extension enabled
- Standard PHP-FPM or mod_php setup
- Document root pointing to the website directory

**Minimal nginx change required:**

Add this one location block to route BlueMap requests to the proxy:

```nginx
location ~ ^/([^/]+)/map/ {
    rewrite ^(.*)$ /proxy.php last;
}
```

That's it! Everything else uses standard PHP configuration.

See `nginx.conf` for a complete example if needed.

**For Apache:** The included `.htaccess` file handles the routing automatically - no changes needed!

### How Proxying Works

BlueMap requests (`/servername/map/*`) are handled by `proxy.php`:
1. Detects map request from URL path
2. Extracts server name from URL
3. Looks up BlueMap backend from `servers.yml`
4. Uses cURL to proxy the request to the backend
5. Streams response back to client