<?php

require_once __DIR__ . '/config.php';

$requestUri = $_SERVER['REQUEST_URI'];
$requestPath = parse_url($requestUri, PHP_URL_PATH);

if (preg_match('#^/([^/]+)/map(/.*)?$#', $requestPath, $matches)) {
    $serverName = $matches[1];
    $server = getServerByName($serverName);
    $restOfPath = $matches[2] ?? '';

    if (!$server) {
        header("HTTP/1.0 404 Not Found");
        echo "Server not found";
        exit;
    }

    // --- Standard BlueMap redirect ---
    if (isset($server['bluemap_url']) && !empty($server['bluemap_url'])) {
        // For live data, we still need to proxy to the bluemap server
        if (strpos($restOfPath, '/live/') === 0) {
            $url = rtrim($server['bluemap_url'], '/') . $restOfPath;
            header("Location: " . $url);
            exit;
        }
        // Fall through to serve UI and map files locally
    } 
    // --- No BlueMap configured ---
    else {
        header("HTTP/1.0 404 Not Found");
        echo "BlueMap is not configured for this server.";
        exit;
    }

    // --- Serve Local UI & Map Files ---
    $webrootPath = __DIR__ . '/data/' . $serverName . '/bluemap/webroot';
    
    // Sanitize path to prevent directory traversal
    $filePath = realpath($webrootPath . $restOfPath);

    // If the path is empty, default to index.html
    if (empty($restOfPath) || $restOfPath === '/') {
        $filePath = $webrootPath . '/index.html';
    }

    if ($filePath && strpos($filePath, $webrootPath) === 0 && file_exists($filePath) && is_file($filePath)) {
        // Serve the file
        $mimeType = mime_content_type($filePath);
        header('Content-Type: ' . $mimeType);
        header('Content-Length: ' . filesize($filePath));
        readfile($filePath);
        exit;
    } else {
        // If the specific file is not found, try to find an index.html in the directory
        $indexPath = realpath($webrootPath . $restOfPath . '/index.html');
        if ($indexPath && strpos($indexPath, $webrootPath) === 0 && file_exists($indexPath) && is_file($indexPath)) {
            $mimeType = mime_content_type($indexPath);
            header('Content-Type: ' . $mimeType);
            header('Content-Length: ' . filesize($indexPath));
            readfile($indexPath);
            exit;
        }

        header("HTTP/1.0 404 Not Found");
        echo "File not found.";
        exit;
    }
} else {
    header("HTTP/1.0 400 Bad Request");
    echo "Invalid request";
    exit;
}
