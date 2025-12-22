<?php
require_once __DIR__ . '/config.php';

$serverName = $_GET['name'] ?? '';
$server = getServerByName($serverName);

if (!$server) {
    header('HTTP/1.0 404 Not Found');
    echo "Server not found";
    exit;
}

$mods = scanModsDirectory($serverName, $server['current_version_path'] ?? '');
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" href="/favicon.ico" type="image/x-icon">
    <title><?php echo htmlspecialchars($server['name']); ?> - <?php echo SITE_TITLE; ?></title>
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
            padding: 20px;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            padding: 40px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
        }
        .server-header {
            text-align: center;
            margin-bottom: 30px;
        }
        .server-header h1 {
            color: #333;
            font-size: 2.5em;
            margin-bottom: 10px;
            text-transform: capitalize;
        }
        .server-status {
            display: inline-block;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 0.9em;
            font-weight: bold;
            margin: 10px 0;
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
        .subtitle {
            color: #666;
            font-size: 1.1em;
        }
        .button-group {
            display: flex;
            gap: 15px;
            margin: 30px 0;
            flex-wrap: wrap;
        }
        .btn {
            flex: 1;
            min-width: 150px;
            padding: 15px 25px;
            text-align: center;
            text-decoration: none;
            color: white;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 8px;
            font-weight: bold;
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
        }
        .btn.disabled {
            background: linear-gradient(135deg, #9ca3af 0%, #6b7280 100%);
            cursor: not-allowed;
            opacity: 0.6;
            pointer-events: none;
        }
        .btn.disabled:hover {
            transform: none;
            box-shadow: none;
        }
        .btn-secondary {
            background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
        }
        .info-box {
            background: #f3f4f6;
            border-radius: 8px;
            padding: 20px;
            margin-top: 30px;
            text-align: left;
        }
        .info-box h3 {
            color: #333;
            margin-bottom: 15px;
            font-size: 1.3em;
        }
        .info-box p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 10px;
        }
        .server-address {
            background: #e5e7eb;
            padding: 8px 12px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            display: inline-block;
            margin-top: 5px;
        }
        .copy-btn {
            background: none;
            border: none;
            cursor: pointer;
            font-size: 1.2em;
            vertical-align: middle;
            margin-left: 5px;
            padding: 5px;
            border-radius: 4px;
            transition: background 0.2s;
        }
        .copy-btn:hover {
            background: rgba(0,0,0,0.1);
        }
        .mods-section {
            margin-top: 20px;
        }
        .mods-section h4 {
            color: #ffff55;
            margin-bottom: 15px;
            font-size: 0.85em;
            text-shadow: 2px 2px 0px rgba(0, 0, 0, 0.5);
            line-height: 1.6;
        }
        .file-tree {
            list-style: none;
            padding: 0;
            margin: 0;
        }
        .file-tree ul {
            list-style: none;
            padding-left: 20px;
            margin: 0;
        }
        .tree-folder,
        .tree-file,
        .tree-link {
            margin: 5px 0;
        }
        .folder-toggle {
            cursor: pointer;
            padding: 8px 12px;
            background: transparent;
            border: none;
            box-shadow: none;
            display: inline-block;
            transition: all 0.1s ease;
            font-size: 0.75em;
            line-height: 1.6;
            color: #ffffff;
            text-shadow: 1px 1px 0px rgba(0, 0, 0, 0.5);
            width: 100%;
            text-align: left;
        }
        .folder-toggle:hover {
            background: rgba(255, 255, 255, 0.1);
            transform: none;
            box-shadow: none;
        }
        .folder-toggle:active {
            transform: none;
            box-shadow: none;
        }
        .folder-icon {
            display: inline-block;
            transition: transform 0.2s;
            margin-right: 5px;
        }
        .folder-collapsed .folder-icon {
            transform: rotate(-90deg);
        }
        .folder-content {
            margin-top: 0;
            padding-left: 20px;
            border-left: 1px dashed rgba(255, 255, 255, 0.3);
            max-height: 2000px;
            overflow: hidden;
            transition: max-height 0.3s ease;
        }
        .folder-content.collapsed {
            max-height: 0;
        }
        .file-link {
            color: #0099ff;
            text-decoration: none;
            font-size: 0.65em;
            line-height: 1.8;
            text-shadow: 1px 1px 0px rgba(0, 0, 0, 0.3);
            padding: 4px 8px;
            background: transparent;
            border: none;
            box-shadow: none;
            display: block;
            transition: all 0.1s ease;
        }
        .file-link:hover {
            color: #33aaff;
            background: rgba(255, 255, 255, 0.1);
            transform: none;
            box-shadow: none;
        }
        .file-size-inline {
            color: #2a2a2a;
            font-size: 0.9em;
            margin-left: 8px;
            float: right;
        }
        .file-icon,
        .link-icon {
            display: inline-block;
            margin-right: 5px;
        }
        .current-version {
            position: relative;
        }
        .tree-folder.current-version > .folder-toggle {
            background: rgba(85, 255, 85, 0.2);
            border: 1px solid #55ff55;
        }
        .current-badge {
            background: #55ff55;
            color: #003300;
            padding: 2px 6px;
            border: 2px solid #000;
            box-shadow: 2px 2px 0px rgba(0, 0, 0, 0.5);
            font-size: 0.8em;
            margin-left: 8px;
            display: inline-block;
            float: right;
        }
        footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e5e7eb;
            color: #666;
            font-size: 0.9em;
            text-align: center;
        }
        footer a {
            color: #667eea;
            text-decoration: none;
        }
        footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="server-header">
            <h1><?php echo htmlspecialchars($server['name']); ?></h1>
            <?php if (!empty($server['status'])): ?>
            <span class="server-status <?php echo getStatusBadgeClass($server['status']); ?>">
                <?php echo htmlspecialchars($server['status']); ?>
            </span>
            <?php endif; ?>
            <?php if (!empty($server['description'])): ?>
            <p class="subtitle"><?php echo htmlspecialchars($server['description']); ?></p>
            <?php endif; ?>
        </div>
        
        <div class="button-group">
            <?php 
            $isOnline = !empty($server['status']) && strtolower($server['status']) === 'online';
            ?>
            <a href="<?php echo $isOnline ? '/' . htmlspecialchars($server['name']) . '/map/' : '#'; ?>" 
               class="btn<?php echo $isOnline ? '' : ' disabled'; ?>">
                🗺️ View Interactive Map
            </a>
            <a href="/" class="btn btn-secondary">← Back to Server List</a>
        </div>
        
        <?php if (!empty($server['description'])): ?>
        <div class="info-box">
            <h3>About <?php echo htmlspecialchars(ucfirst($server['name'])); ?></h3>
            <p><?php echo htmlspecialchars($server['description']); ?></p>
        </div>
        <?php endif; ?>
        
        <?php if (!empty($server['server_address']) || !empty($server['minecraft_version']) || !empty($server['modloader']) || !empty($mods['files']) || !empty($mods['links'])): ?>
        <div class="info-box">
            <h3>Server Information</h3>
            
            <?php if (!empty($server['server_address'])): ?>
            <p>
                <strong>🌐 Server Address:</strong><br>
                <span class="server-address" id="server-ip"><?php echo htmlspecialchars($server['server_address']); ?></span>
                <button class="copy-btn" onclick="copyToClipboard('server-ip', this)" title="Copy IP">📋</button>
            </p>
            <?php endif; ?>

            <?php if (!empty($server['automodpack_fingerprint'])): ?>
            <p>
                <strong>📦 AutoModPack Fingerprint:</strong><br>
                <span class="server-address" id="automodpack-fingerprint"><?php echo htmlspecialchars($server['automodpack_fingerprint']); ?></span>
                <button class="copy-btn" onclick="copyToClipboard('automodpack-fingerprint', this)" title="Copy Fingerprint">📋</button>
            </p>
            <?php endif; ?>
            
            <?php 
            $serverInfo = getPlayerInfo($server);
            if ($serverInfo): 
            ?>
            <p style="margin-top: 10px;">
                <strong>👥 Players:</strong>
                <span id="player-count"><?php echo $serverInfo['players']['online']; ?> / <?php echo $serverInfo['players']['max']; ?></span>
            </p>
            <?php if (!empty($serverInfo['players']['sample'])): ?>
            <p style="margin-top: 5px; font-size: 0.9em; color: #555;">
                <strong>📝 Online:</strong>
                <span id="player-list">
                <?php 
                $playerNames = array_map(function($p) { return $p['name']; }, $serverInfo['players']['sample']);
                echo htmlspecialchars(implode(', ', $playerNames));
                ?>
                </span>
            </p>
            <?php endif; ?>
            
            <?php if (!empty($serverInfo['software'])): ?>
            <p style="margin-top: 5px;">
                <strong>💻 Software:</strong>
                <?php echo htmlspecialchars($serverInfo['software']); ?>
            </p>
            <?php endif; ?>

            <?php if (!empty($serverInfo['plugins'])): ?>
            <p style="margin-top: 5px;">
                <strong>🔌 Plugins:</strong>
                <?php 
                $plugins = is_array($serverInfo['plugins']) ? implode(', ', $serverInfo['plugins']) : $serverInfo['plugins'];
                echo htmlspecialchars($plugins); 
                ?>
            </p>
            <?php endif; ?>

            <?php endif; ?>
            
            <?php if (!empty($server['minecraft_version'])): ?>
            <p>
                <strong>🎮 Minecraft Version:</strong>
                <?php echo htmlspecialchars($server['minecraft_version']); ?>
            </p>
            <?php endif; ?>
            
            <?php if (!empty($server['modloader'])): ?>
            <p>
                <strong>🔧 Modloader:</strong>
                <?php echo htmlspecialchars(ucfirst($server['modloader'])); ?>
            </p>
            <?php endif; ?>
            
            <?php if (!empty($mods['tree'])): ?>
            <div class="mods-section">
                <h4>� Downloads:</h4>
                <?php 
                function renderTree($items, $serverName, $currentVersionPath = '') {
                    if (empty($items)) return;
                    echo '<ul class="file-tree">';
                    foreach ($items as $item) {
                        $isCurrent = !empty($currentVersionPath) && $item['path'] === $currentVersionPath;
                        $currentClass = $isCurrent ? ' current-version' : '';
                        
                        if ($item['type'] === 'folder') {
                            $folderId = 'folder-' . md5($item['path']);
                            
                            // Determine if this folder should be expanded
                            // Expand if it is the current version or a parent of the current version
                            $isParent = !empty($currentVersionPath) && strpos($currentVersionPath, $item['path'] . '/') === 0;
                            $shouldExpand = $isCurrent || $isParent;
                            
                            $collapsedClass = $shouldExpand ? '' : ' collapsed';
                            $folderCollapsedClass = $shouldExpand ? '' : ' folder-collapsed';
                            
                            echo '<li class="tree-folder' . $currentClass . $folderCollapsedClass . '">';
                            echo '<span class="folder-toggle" onclick="toggleFolder(\'' . $folderId . '\')">';
                            echo '<span class="folder-icon">📁</span> ';
                            echo htmlspecialchars($item['name']);
                            if ($isCurrent) echo ' <span class="current-badge">Current</span>';
                            echo '</span>';
                            echo '<div class="folder-content' . $collapsedClass . '" id="' . $folderId . '">';
                            renderTree($item['children'], $serverName, $currentVersionPath);
                            echo '</div>';
                            echo '</li>';
                        } elseif ($item['type'] === 'file') {
                            $isMarkdown = preg_match('/\.(md|markdown)$/i', $item['name']);
                            $downloadAttr = $isMarkdown ? '' : ' download="' . htmlspecialchars($item['name']) . '"';
                            $icon = $isMarkdown ? '📄' : '📦';

                            echo '<li class="tree-file' . $currentClass . '">';
                            echo '<a href="/' . htmlspecialchars($serverName) . '/mods/' . htmlspecialchars($item['path']) . '" ';
                            echo 'class="file-link"' . $downloadAttr . '>';
                            echo '<span class="file-icon">' . $icon . '</span> ';
                            echo htmlspecialchars($item['name']);
                            echo ' <span class="file-size-inline">' . formatFileSize($item['size']) . '</span>';
                            echo '</a>';
                            echo '</li>';
                        } elseif ($item['type'] === 'link') {
                            echo '<li class="tree-link' . $currentClass . '">';
                            echo '<a href="' . htmlspecialchars($item['url']) . '" class="file-link" target="_blank">';
                            echo '<span class="link-icon">🔗</span> ';
                            echo htmlspecialchars($item['name']);
                            echo '</a>';
                            echo '</li>';
                        }
                    }
                    echo '</ul>';
                }
                renderTree($mods['tree'], $server['name'], $server['current_version_path'] ?? '');
                ?>
            </div>
            <?php endif; ?>
        </div>
        <?php endif; ?>
        
                <footer>
            <p>
                <a href="/">All Servers</a> | 
                <a href="<?php echo PTERODACTYL_PANEL_URL; ?>" target="_blank">Admin Panel</a>
            </p>
            <p style="margin-top: 10px;">
                <a href="https://wallpapers.com/wallpapers/moving-minecraft-steve-with-a-horse-pdsh695zjw7xj33p.html" target="_blank" rel="noopener">Wallpaper by fashions</a> on Wallpapers.com
            </p>
        </footer>
    </div>
    
    <script>
    function copyToClipboard(elementId, button) {
        const text = document.getElementById(elementId).innerText;
        navigator.clipboard.writeText(text).then(() => {
            const originalText = button.innerText;
            button.innerText = '✅';
            setTimeout(() => {
                button.innerText = originalText;
            }, 2000);
        }).catch(err => {
            console.error('Failed to copy: ', err);
            alert('Failed to copy to clipboard');
        });
    }

    function toggleFolder(folderId) {
        const folderContent = document.getElementById(folderId);
        const folderElement = folderContent.parentElement;
        
        if (folderContent.classList.contains('collapsed')) {
            folderContent.classList.remove('collapsed');
            folderElement.classList.remove('folder-collapsed');
        } else {
            folderContent.classList.add('collapsed');
            folderElement.classList.add('folder-collapsed');
        }
    }

    <?php if ($isOnline && !empty($server['bluemap_proxy'])): ?>
    function fetchBlueMapPlayers() {
        const bluemapUrl = '/<?php echo htmlspecialchars($server['name']); ?>/map/maps/world/live/players.json';
        fetch(bluemapUrl)
            .then(response => {
                if (!response.ok) {
                    throw new Error('BlueMap data not available');
                }
                return response.json();
            })
            .then(data => {
                if (data && data.players) {
                    const playerCountSpan = document.getElementById('player-count');
                    const playerListSpan = document.getElementById('player-list');
                    
                    if (playerCountSpan) {
                        const maxPlayers = playerCountSpan.innerText.split('/')[1].trim();
                        playerCountSpan.innerText = `${data.players.length} / ${maxPlayers}`;
                    }
                    
                    if (playerListSpan) {
                        const playerNames = data.players.map(p => p.name).join(', ');
                        playerListSpan.innerText = playerNames || 'No players online';
                    }
                }
            })
            .catch(error => {
                console.warn('Could not fetch BlueMap player list:', error);
            });
    }

    document.addEventListener('DOMContentLoaded', () => {
        fetchBlueMapPlayers();
        setInterval(fetchBlueMapPlayers, 10000); // Refresh every 10 seconds
    });
    <?php endif; ?>
    </script>
</body>
</html>
