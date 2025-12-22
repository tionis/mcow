<?php
/**
 * Minecraft Server Website Configuration
 */

require_once __DIR__ . '/deps/mc_query/src/MinecraftPing.php';
require_once __DIR__ . '/deps/mc_query/src/MinecraftPingException.php';
require_once __DIR__ . '/deps/mc_query/src/MinecraftQuery.php';
require_once __DIR__ . '/deps/mc_query/src/MinecraftQueryException.php';

use xPaw\MinecraftQuery;
use xPaw\MinecraftQueryException;
use xPaw\MinecraftPing;

// Load Symfony YAML parser (if available) or use simple parser
function loadServersConfig($configFile = __DIR__ . '/servers.yml') {
    if (!file_exists($configFile)) {
        return ['servers' => []];
    }
    
    // Try to use symfony/yaml if available
    if (class_exists('Symfony\Component\Yaml\Yaml')) {
        return \Symfony\Component\Yaml\Yaml::parseFile($configFile);
    }
    
    // Fallback to simple YAML parser
    return parseSimpleYaml(file_get_contents($configFile));
}

// Simple YAML parser for basic structures
function parseSimpleYaml($yaml) {
    $lines = explode("\n", $yaml);
    $result = ['servers' => []];
    $currentServer = null;
    $indent = 0;
    
    foreach ($lines as $line) {
        $trimmed = trim($line);
        if (empty($trimmed) || $trimmed[0] === '#') continue;
        
        // Detect list item
        if (preg_match('/^- name: ["\']?([^"\']+)["\']?/', $trimmed, $matches)) {
            if ($currentServer) {
                $result['servers'][] = $currentServer;
            }
            $currentServer = ['name' => $matches[1]];
        } elseif ($currentServer && preg_match('/^([a-z_]+): ["\']?([^"\']+)["\']?/', $trimmed, $matches)) {
            $currentServer[$matches[1]] = $matches[2];
        }
    }
    
    if ($currentServer) {
        $result['servers'][] = $currentServer;
    }
    
    return $result;
}

function getServers() {
    $config = loadServersConfig();
    return $config['servers'] ?? [];
}

function getPlayerInfo($server, $fetchPlayerNames = true) {
    // This function now primarily uses the direct server query, which is faster.
    // The BlueMap player list fetching is offloaded to client-side JavaScript.
    $forcePing = !$fetchPlayerNames;
    return queryServer($server, $forcePing);
}

/**
 * Queries a server for its status.
 * Uses xPaw's MinecraftQuery library.
 * @param array $server The server configuration array.
 * @param bool $forcePing If true, only a lightweight ping will be used.
 * @return array|false The server information or false on failure.
 */
function queryServer($server, $forcePing = false) {
    if (empty($server['server_address']) || strtolower($server['status']) !== 'online') {
        return false;
    }

    $address = $server['server_address'];
    $host = $address;
    $port = 25565;

    if (strpos($address, ':') !== false) {
        $parts = explode(':', $address);
        $host = $parts[0];
        $port = intval($parts[1]);
    }

    try {
        // Use full query if enabled and not forced to ping
        if (!$forcePing && !empty($server['enable_query'])) {
            $Query = new MinecraftQuery();
            $Query->Connect( $host, $port );
            $Info = $Query->GetInfo();
            $Players = $Query->GetPlayers();
            
            if($Info) {
                return [
                    'version' => [
                        'name' => $Info['Version'],
                        'protocol' => 0
                    ],
                    'players' => [
                        'max' => $Info['MaxPlayers'],
                        'online' => $Info['Players'],
                        'sample' => $Players ? array_map(function($p) { return ['name' => $p]; }, $Players) : []
                    ],
                    'description' => $Info['HostName'],
                    'favicon' => '', // Query doesn't return favicon
                    'software' => $Info['Software'],
                    'plugins' => $Info['Plugins']
                ];
            }
        }

        // Fallback to ping (or if ping was forced)
        // The constructor for MinecraftPing attempts connection and can throw an exception.
        $QueryPing = new MinecraftPing( $host, $port, 1 );
        $PingInfo = $QueryPing->Query();
        
        if($PingInfo) {
            return [
                'version' => [
                    'name' => $PingInfo['version']['name'] ?? '',
                    'protocol' => 0
                ],
                'players' => [
                    'max' => $PingInfo['players']['max'] ?? 0,
                    'online' => $PingInfo['players']['online'] ?? 0,
                    'sample' => $PingInfo['players']['sample'] ?? []
                ],
                'description' => $PingInfo['description'] ?? '',
                'favicon' => '', // Ping doesn't return favicon
                'software' => null,
                'plugins' => null
            ];
        }
    } catch (MinecraftQueryException | MinecraftPingException $e) {
        // Server is offline or query failed.
        // We can log the error here if needed, for now, we'll just return false.
    }

    return false;
}

function getServerByName($name) {
    $servers = getServers();
    foreach ($servers as $server) {
        if ($server['name'] === $name) {
            return $server;
        }
    }
    return null;
}

function getStatusBadgeClass($status) {
    $status = strtolower($status);
    switch ($status) {
        case 'online':
            return 'status-online';
        case 'planned':
            return 'status-planned';
        case 'offline':
        default:
            return 'status-offline';
    }
}

function scanModsDirectory($serverName, $currentVersionPath = '') {
    // Look for mods in the data directory to avoid URL collisions
    $modsDir = __DIR__ . '/data/' . $serverName . '/mods';
    if (!is_dir($modsDir)) {
        return ['files' => [], 'links' => [], 'tree' => []];
    }
    
    $result = ['files' => [], 'links' => [], 'tree' => []];
    $items = new RecursiveIteratorIterator(
        new RecursiveDirectoryIterator($modsDir, RecursiveDirectoryIterator::SKIP_DOTS),
        RecursiveIteratorIterator::SELF_FIRST
    );
    
    foreach ($items as $item) {
        if ($item->isFile()) {
            $relativePath = str_replace($modsDir . '/', '', $item->getPathname());
            
            if (substr($item->getFilename(), -4) === '.url') {
                // Parse .url file
                $content = file_get_contents($item->getPathname());
                if (preg_match('/URL=(.+)$/m', $content, $matches)) {
                    $result['links'][] = [
                        'name' => substr($item->getFilename(), 0, -4),
                        'url' => trim($matches[1]),
                        'path' => $relativePath
                    ];
                }
            } else {
                // Regular file
                $result['files'][] = [
                    'name' => $item->getFilename(),
                    'path' => $relativePath,
                    'size' => $item->getSize(),
                    'url' => '/' . $serverName . '/mods/' . $relativePath
                ];
            }
        }
    }
    
    // Build tree structure
    $result['tree'] = buildDirectoryTree($modsDir, $serverName, '', $currentVersionPath);
    
    return $result;
}

function buildDirectoryTree($dir, $serverName, $relativePath = '', $currentVersionPath = '') {
    $tree = [];
    
    if (!is_dir($dir)) {
        return $tree;
    }
    
    $items = scandir($dir);
    
    foreach ($items as $item) {
        if ($item === '.' || $item === '..') {
            continue;
        }
        
        $fullPath = $dir . '/' . $item;
        $relPath = $relativePath ? $relativePath . '/' . $item : $item;
        
        if (is_dir($fullPath)) {
            $tree[] = [
                'type' => 'folder',
                'name' => $item,
                'path' => $relPath,
                'children' => buildDirectoryTree($fullPath, $serverName, $relPath, $currentVersionPath)
            ];
        } else {
            if (substr($item, -4) === '.url') {
                // Parse .url file
                $content = file_get_contents($fullPath);
                if (preg_match('/URL=(.+)$/m', $content, $matches)) {
                    $tree[] = [
                        'type' => 'link',
                        'name' => substr($item, 0, -4),
                        'path' => $relPath,
                        'url' => trim($matches[1])
                    ];
                }
            } else {
                $tree[] = [
                    'type' => 'file',
                    'name' => $item,
                    'path' => $relPath,
                    'size' => filesize($fullPath),
                    'url' => '/' . $serverName . '/mods/' . $relPath
                ];
            }
        }
    }

    // Sort tree: Current version first, then folders, then files
    usort($tree, function($a, $b) use ($currentVersionPath) {
        // 1. Current version path (exact match)
        $aIsCurrent = ($a['path'] === $currentVersionPath);
        $bIsCurrent = ($b['path'] === $currentVersionPath);
        if ($aIsCurrent && !$bIsCurrent) return -1;
        if (!$aIsCurrent && $bIsCurrent) return 1;

        // 2. Parent of current version path
        $aIsParent = $currentVersionPath && strpos($currentVersionPath, $a['path'] . '/') === 0;
        $bIsParent = $currentVersionPath && strpos($currentVersionPath, $b['path'] . '/') === 0;
        if ($aIsParent && !$bIsParent) return -1;
        if (!$aIsParent && $bIsParent) return 1;

        // 3. Folders before files
        if ($a['type'] === 'folder' && $b['type'] !== 'folder') return -1;
        if ($a['type'] !== 'folder' && $b['type'] === 'folder') return 1;

        // 4. Alphabetical
        return strcasecmp($a['name'], $b['name']);
    });
    
    return $tree;
}

function formatFileSize($bytes) {
    $units = ['B', 'KB', 'MB', 'GB'];
    $bytes = max($bytes, 0);
    $pow = floor(($bytes ? log($bytes) : 0) / log(1024));
    $pow = min($pow, count($units) - 1);
    $bytes /= (1 << (10 * $pow));
    return round($bytes, 2) . ' ' . $units[$pow];
}

// Configuration constants
define('PTERODACTYL_PANEL_URL', 'https://panel.mc.ieee-passau.org');
define('SITE_TITLE', 'IEEE SB Passau Minecraft Servers');
?>
