<?php
require_once __DIR__ . '/config.php';

/**
 * Router Logic
 */
$requestUri = $_SERVER['REQUEST_URI'];
$requestPath = parse_url($requestUri, PHP_URL_PATH);

// Check if it's a BlueMap proxy request
if (preg_match('#^/([^/]+)/map(/.*)?$#', $requestPath)) {
    require __DIR__ . '/bluemap.php';
    exit;
}

// Check if it's a mod download request
if (preg_match('#^/([^/]+)/mods/(.+)$#', $requestPath, $matches)) {
    $serverName = $matches[1];
    $filePath = urldecode($matches[2]);
    
    // Security check: prevent directory traversal
    if (strpos($filePath, '..') !== false) {
        header('HTTP/1.0 403 Forbidden');
        echo "Access denied";
        exit;
    }
    
    // Look in data directory
    $fullPath = __DIR__ . '/data/' . $serverName . '/mods/' . $filePath;
    
    if (file_exists($fullPath) && is_file($fullPath)) {
        // Check if it's a markdown file
        $extension = strtolower(pathinfo($fullPath, PATHINFO_EXTENSION));
        if ($extension === 'md' || $extension === 'markdown') {
            require_once __DIR__ . '/deps/parsedown/Parsedown.php';
            $Parsedown = new Parsedown();
            $content = file_get_contents($fullPath);
            $html = $Parsedown->text($content);
            
            // Render page with markdown content
            $server = getServerByName($serverName);
            ?>
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <link rel="icon" href="/favicon.ico" type="image/x-icon">
                <title><?php echo htmlspecialchars(basename($filePath)); ?> - <?php echo htmlspecialchars($server['name']); ?></title>
                <style>
                    body {
                        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
                        background: url('/background.jpg') no-repeat center center fixed;
                        background-size: cover;
                        min-height: 100vh;
                        padding: 20px;
                        margin: 0;
                    }
                    .container {
                        max-width: 900px;
                        margin: 0 auto;
                        background: white;
                        border-radius: 12px;
                        padding: 40px;
                        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
                    }
                    .markdown-body {
                        line-height: 1.6;
                        color: #333;
                    }
                    .markdown-body h1, .markdown-body h2, .markdown-body h3 { margin-top: 24px; margin-bottom: 16px; font-weight: 600; line-height: 1.25; }
                    .markdown-body h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
                    .markdown-body h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
                    .markdown-body p { margin-top: 0; margin-bottom: 16px; }
                    .markdown-body code { background-color: rgba(27,31,35,.05); border-radius: 3px; font-size: 85%; margin: 0; padding: .2em .4em; }
                    .markdown-body pre { background-color: #f6f8fa; border-radius: 3px; font-size: 85%; line-height: 1.45; overflow: auto; padding: 16px; }
                    .markdown-body pre code { background-color: transparent; border: 0; display: inline; line-height: inherit; margin: 0; overflow: visible; padding: 0; word-wrap: normal; }
                    .markdown-body blockquote { border-left: .25em solid #dfe2e5; color: #6a737d; padding: 0 1em; margin: 0; }
                    .markdown-body ul, .markdown-body ol { padding-left: 2em; }
                    .btn {
                        display: inline-block;
                        padding: 10px 20px;
                        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                        color: white;
                        text-decoration: none;
                        border-radius: 6px;
                        font-weight: bold;
                        margin-bottom: 20px;
                    }
                    .btn:hover { transform: translateY(-2px); box-shadow: 0 5px 15px rgba(0,0,0,0.1); }
                </style>
            </head>
            <body>
                <div class="container">
                    <a href="/<?php echo htmlspecialchars($serverName); ?>/" class="btn">← Back to Server</a>
                    <div class="markdown-body">
                        <?php echo $html; ?>
                    </div>
                </div>
            </body>
            </html>
            <?php
            exit;
        }

        // Serve the file
        $mimeType = mime_content_type($fullPath);
        header('Content-Type: ' . $mimeType);
        header('Content-Disposition: attachment; filename="' . basename($fullPath) . '"');
        header('Content-Length: ' . filesize($fullPath));
        readfile($fullPath);
        exit;
    } else {
        header('HTTP/1.0 404 Not Found');
        echo "File not found";
        exit;
    }
}

// Check if requesting a specific server's page directly (new style)
if (preg_match('#^/([^/]+)/?$#', $requestPath, $matches)) {
    $serverName = $matches[1];
    // Check if this server exists
    if (getServerByName($serverName)) {
        $_GET['name'] = $serverName;
        require __DIR__ . '/server.php';
        exit;
    }
}

$servers = getServers();
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
    <title><?php echo SITE_TITLE; ?></title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: url('/background.jpg') no-repeat center center fixed;
            background-size: cover;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            width: 100%;
        }
        h1 {
            text-align: center;
            color: white;
            font-size: 3em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
        }
        .subtitle {
            text-align: center;
            color: rgba(255, 255, 255, 0.9);
            font-size: 1.2em;
            margin-bottom: 40px;
        }
        .servers-grid {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            gap: 30px;
            margin-bottom: 40px;
        }
        .server-card {
            flex: 1 1 300px;
            max-width: 450px;
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
            cursor: pointer;
            position: relative;
        }
        .server-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 40px rgba(0, 0, 0, 0.3);
        }
        .server-name {
            font-size: 1.8em;
            font-weight: bold;
            margin-bottom: 10px;
            color: #333;
            text-transform: capitalize;
        }
        .server-description {
            color: #666;
            margin-bottom: 15px;
            line-height: 1.5;
        }
        .server-metadata {
            color: #888;
            font-size: 0.9em;
            margin-bottom: 15px;
            line-height: 1.6;
        }
        .server-metadata strong {
            color: #555;
        }
        .server-status {
            display: inline-block;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: bold;
            margin-bottom: 15px;
        }
        .status-online {
            background: #10b981;
            color: white;
        }
        .status-planned {
            background: #f59e0b;
            color: white;
        }
        .status-offline {
            background: #ef4444;
            color: white;
        }
        .button-group {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        .btn {
            flex: 1;
            min-width: 120px;
            padding: 12px 20px;
            text-align: center;
            text-decoration: none;
            color: white;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 8px;
            font-weight: bold;
            transition: transform 0.3s ease, box-shadow 0.3s ease;
            border: none;
            cursor: pointer;
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
        }
        .btn:disabled,
        .btn.disabled {
            background: linear-gradient(135deg, #9ca3af 0%, #6b7280 100%);
            cursor: not-allowed;
            opacity: 0.6;
        }
        .btn:disabled:hover,
        .btn.disabled:hover {
            transform: none;
            box-shadow: none;
        }
        footer {
            text-align: center;
            color: white;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid rgba(255, 255, 255, 0.3);
        }
        footer a {
            color: white;
            text-decoration: none;
            font-weight: bold;
        }
        footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>y
    <div class="container">
        <h1><?php echo SITE_TITLE; ?></h1>
        <!-- <p class="subtitle">Minecraft Servers</p> -->
        
        <div class="servers-grid">
            <?php foreach ($servers as $server): ?>
            <div class="server-card" onclick="window.location.href='/<?php echo urlencode($server['name']); ?>/'">
                <div class="server-name"><?php echo htmlspecialchars($server['name']); ?></div>
                <?php if (!empty($server['status'])): ?>
                <span class="server-status <?php echo getStatusBadgeClass($server['status']); ?>">
                    <?php echo htmlspecialchars($server['status']); ?>
                </span>
                <?php endif; ?>

                <?php 
                $serverInfo = getPlayerInfo($server, false);
                if ($serverInfo): 
                ?>
                <span class="server-status" style="background: #3b82f6; color: white; margin-left: 5px;">
                    👥 <?php echo $serverInfo['players']['online']; ?>/<?php echo $serverInfo['players']['max']; ?>
                </span>
                <?php endif; ?>

                <?php if (!empty($server['description'])): ?>
                <p class="server-description"><?php echo htmlspecialchars($server['description']); ?></p>
                <?php endif; ?>
                <?php if (!empty($server['minecraft_version']) || !empty($server['modloader'])): ?>
                <div class="server-metadata">
                    <?php if (!empty($server['minecraft_version'])): ?>
                    <div><strong>MC Version:</strong> <?php echo htmlspecialchars($server['minecraft_version']); ?></div>
                    <?php endif; ?>
                    <?php if (!empty($server['modloader'])): ?>
                    <div><strong>Modloader:</strong> <?php echo htmlspecialchars(ucfirst($server['modloader'])); ?></div>
                    <?php endif; ?>
                </div>
                <?php endif; ?>
                <div class="button-group">
                    <?php 
                    $isOnline = !empty($server['status']) && strtolower($server['status']) === 'online';
                    ?>
                    <button class="btn" 
                            <?php if (!$isOnline): ?>disabled<?php endif; ?>
                            onclick="<?php if ($isOnline): ?>event.stopPropagation(); window.location.href='/<?php echo htmlspecialchars($server['name']); ?>/map/'<?php endif; ?>">
                        🗺️ View Map
                    </button>
                    <button class="btn" onclick="event.stopPropagation(); window.location.href='/<?php echo urlencode($server['name']); ?>/'">
                        📋 Details
                    </button>
                </div>
            </div>
            <?php endforeach; ?>
        </div>
        
        <footer>
            <p>
                <a href="<?php echo PTERODACTYL_PANEL_URL; ?>" target="_blank">Admin Panel</a>
            </p>
            <p style="margin-top: 10px;">
                <a href="https://wallpapers.com/wallpapers/moving-minecraft-steve-with-a-horse-pdsh695zjw7xj33p.html" target="_blank" rel="noopener">Wallpaper by fashions</a> on Wallpapers.com
            </p>
        </footer>
    </div>
</body>
</html>
