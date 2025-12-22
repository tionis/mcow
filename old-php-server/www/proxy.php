<?php
/**
 * BlueMap Proxy
 * Proxies requests to BlueMap backends based on server name
 */

require_once __DIR__ . '/config.php';

// Get the request path
$requestUri = $_SERVER['REQUEST_URI'];

// Extract server name from path: /servername/map/...
if (!preg_match('#^/([^/]+)/map(/.*)?$#', $requestUri, $matches)) {
    http_response_code(404);
    die('Invalid proxy path');
}

$serverName = $matches[1];
$mapPath = $matches[2] ?? '/';

// Get server configuration
$server = getServerByName($serverName);
if (!$server || empty($server['bluemap_proxy'])) {
    http_response_code(404);
    die('Server not found or BlueMap not configured');
}

$blueMapBackend = $server['bluemap_proxy'];

// Build target URL
$targetUrl = 'http://' . $blueMapBackend . $mapPath;
if (!empty($_SERVER['QUERY_STRING'])) {
    $targetUrl .= '?' . $_SERVER['QUERY_STRING'];
}

// Initialize cURL
$ch = curl_init($targetUrl);

// Set cURL options
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, false);
curl_setopt($ch, CURLOPT_HEADER, false);
curl_setopt($ch, CURLOPT_BINARYTRANSFER, true);
curl_setopt($ch, CURLOPT_TIMEOUT, 300); // 5 minute timeout
curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 10); // 10 second connection timeout
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false); // In case BlueMap uses self-signed cert

// Forward request headers
$requestHeaders = [];
foreach (getallheaders() as $name => $value) {
    // Skip host header, we'll set it to the backend
    if (strtolower($name) === 'host') {
        continue;
    }
    $requestHeaders[] = "$name: $value";
}
// Set backend host
$requestHeaders[] = "Host: " . $blueMapBackend;
curl_setopt($ch, CURLOPT_HTTPHEADER, $requestHeaders);

// Handle request method
$method = $_SERVER['REQUEST_METHOD'];
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);

// Forward request body for POST/PUT
if (in_array($method, ['POST', 'PUT', 'PATCH'])) {
    $body = file_get_contents('php://input');
    curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
}

// Stream response headers
curl_setopt($ch, CURLOPT_HEADERFUNCTION, function($curl, $header) {
    $len = strlen($header);
    $header = trim($header);
    
    if (empty($header)) {
        return $len;
    }
    
    // Parse header
    $parts = explode(':', $header, 2);
    if (count($parts) == 2) {
        $name = trim($parts[0]);
        $value = trim($parts[1]);
        
        // Skip certain headers
        $skipHeaders = ['transfer-encoding', 'connection'];
        if (in_array(strtolower($name), $skipHeaders)) {
            return $len;
        }
        
        header("$name: $value", false);
    } elseif (strpos($header, 'HTTP/') === 0) {
        // Status line
        $statusParts = explode(' ', $header, 3);
        if (count($statusParts) >= 2) {
            http_response_code((int)$statusParts[1]);
        }
    }
    
    return $len;
});

// Stream response body directly to output
curl_setopt($ch, CURLOPT_WRITEFUNCTION, function($curl, $data) {
    echo $data;
    flush();
    return strlen($data);
});

// Disable buffering
if (ob_get_level()) {
    ob_end_clean();
}

// Execute request
$success = curl_exec($ch);

if (!$success) {
    $error = curl_error($ch);
    http_response_code(502);
    die("Proxy error: " . htmlspecialchars($error));
}

curl_close($ch);
