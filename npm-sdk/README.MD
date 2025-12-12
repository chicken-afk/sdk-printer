# Printer SDK - JavaScript Client

JavaScript SDK untuk berkomunikasi dengan Printer SDK Server menggunakan **WebSocket** untuk real-time health monitoring.

## 🚀 Features

- ✅ **WebSocket-based Health Check** - Real-time connection monitoring tanpa polling
- ✅ **Auto Reconnection** - Otomatis reconnect jika koneksi terputus
- ✅ **Connection Status Callbacks** - Notifikasi real-time saat koneksi berubah
- ✅ **Heartbeat Monitoring** - Server kirim heartbeat setiap 30 detik
- ✅ **Ping/Pong Support** - Manual ping test ke server
- ✅ **Printer Management** - Get printers, paper sizes, dan print files

## 📦 Installation

### Browser (CDN)
```html
<script src="path/to/index.js"></script>
```

### NPM (Coming Soon)
```bash
npm install @warung-aceh/printer-sdk
```

## 💻 Usage

### Basic Usage

```javascript
// Initialize SDK
const sdk = new PrinterSDK('http://localhost:8971/api/v1');

// Connection change callback
sdk.onConnectionChange = (isConnected, maxRetriesReached) => {
    if (isConnected) {
        console.log('✅ Connected to Printer SDK');
        // Enable UI elements
    } else {
        console.log('❌ Disconnected from Printer SDK');
        if (maxRetriesReached) {
            alert('Failed to connect. Please check if SDK is running.');
        }
    }
};

// Health check callback (heartbeat setiap 30 detik)
sdk.onHealthCheck = (data) => {
    console.log('💓 Health check:', data);
    // data = { status: "online", timestamp: "...", message: "heartbeat" }
};
```

### Check Connection Status

```javascript
// Get current connection status
const status = sdk.getConnectionStatus();
console.log(status);
// {
//   isConnected: true,
//   reconnectAttempts: 0,
//   wsReadyState: 1 // WebSocket.OPEN
// }

// Check if connected
if (await sdk.checkConnection()) {
    console.log('Server is accessible');
}
```

### Get Printers

```javascript
try {
    const printers = await sdk.getPrinters();
    console.log('Available printers:', printers);
    // [
    //   {
    //     id: "abc123",
    //     name: "HP LaserJet Pro",
    //     paper_sizes: ["A4", "Letter", "Legal"]
    //   }
    // ]
} catch (error) {
    console.error('Failed to get printers:', error);
}
```

### Get Paper Sizes

```javascript
const printerId = "abc123";
const papers = await sdk.getPaper(printerId);
console.log('Paper sizes:', papers);
// ["A4", "Letter", "Legal", "A5"]
```

### Print File

```javascript
const printerId = "abc123";
const paperSize = "A4";
const file = document.querySelector('#fileInput').files[0];

try {
    const result = await sdk.print(printerId, paperSize, file);
    console.log('Print successful:', result);
} catch (error) {
    console.error('Print failed:', error.message);
}
```

### Manual Ping

```javascript
// Send ping to server
if (sdk.sendPing()) {
    console.log('Ping sent successfully');
} else {
    console.log('Failed to send ping - not connected');
}

// Server akan respond dengan pong yang diterima di onHealthCheck callback
```

### Stop Health Check

```javascript
// Stop WebSocket connection
sdk.stopHealthCheck();
```

## 🎯 Connection Management

### Auto Reconnection

SDK akan otomatis mencoba reconnect jika koneksi terputus:
- **Max Attempts**: 5 kali
- **Delay**: 3 detik antar percobaan
- **Callback**: Notifikasi via `onConnectionChange`

```javascript
sdk.onConnectionChange = (isConnected, maxRetriesReached) => {
    if (!isConnected) {
        if (maxRetriesReached) {
            // Sudah mencoba 5x, gagal
            console.log('Max retries reached. Please restart SDK.');
        } else {
            // Masih mencoba reconnect
            const status = sdk.getConnectionStatus();
            console.log(`Reconnecting... (${status.reconnectAttempts}/5)`);
        }
    }
};
```

## 📊 WebSocket Events

SDK menerima beberapa tipe message dari server:

### 1. Initial Connection
```json
{
  "status": "connected",
  "timestamp": "2025-12-12T10:30:00Z",
  "message": "SDK Printer Server is running"
}
```

### 2. Heartbeat (Setiap 30 detik)
```json
{
  "status": "online",
  "timestamp": "2025-12-12T10:30:30Z",
  "message": "heartbeat"
}
```

### 3. Pong Response
```json
{
  "status": "online",
  "timestamp": "2025-12-12T10:30:45Z",
  "message": "pong"
}
```

## 🎨 Complete Example

```html
<!DOCTYPE html>
<html>
<head>
    <title>Printer SDK Demo</title>
</head>
<body>
    <div id="status">Disconnected</div>
    <button onclick="loadPrinters()">Load Printers</button>
    <div id="printers"></div>

    <script src="src/index.js"></script>
    <script>
        let sdk = null;

        // Initialize SDK
        function init() {
            sdk = new PrinterSDK('http://localhost:8971/api/v1');

            // Connection callback
            sdk.onConnectionChange = (isConnected, maxRetriesReached) => {
                const statusEl = document.getElementById('status');
                if (isConnected) {
                    statusEl.textContent = '✅ Connected';
                    statusEl.style.color = 'green';
                } else {
                    statusEl.textContent = '❌ Disconnected';
                    statusEl.style.color = 'red';
                    
                    if (maxRetriesReached) {
                        alert('Failed to connect to Printer SDK. Please check if it is running.');
                    }
                }
            };

            // Health check callback
            sdk.onHealthCheck = (data) => {
                console.log('Health:', data);
            };
        }

        // Load printers
        async function loadPrinters() {
            if (!sdk.isConnected) {
                alert('Not connected to SDK');
                return;
            }

            try {
                const printers = await sdk.getPrinters();
                const container = document.getElementById('printers');
                container.innerHTML = printers.map(p => 
                    `<div>${p.name} (${p.id})</div>`
                ).join('');
            } catch (err) {
                alert('Error: ' + err.message);
            }
        }

        // Initialize on load
        window.onload = init;

        // Cleanup on unload
        window.onbeforeunload = () => {
            if (sdk) sdk.stopHealthCheck();
        };
    </script>
</body>
</html>
```

## 🔧 Configuration

### Custom WebSocket URL

```javascript
// Default: auto-converts HTTP to WS
const sdk = new PrinterSDK('http://localhost:8971/api/v1');
// WebSocket URL: ws://localhost:8971/api/v1/ws/connection

// Custom base URL
const sdk = new PrinterSDK('http://192.168.1.100:8971/api/v1');
```

### Customize Reconnection Settings

```javascript
const sdk = new PrinterSDK();

// Customize reconnection
sdk.maxReconnectAttempts = 10; // Default: 5
sdk.reconnectDelay = 5000; // Default: 3000ms (3 seconds)
```

## 🐛 Troubleshooting

### WebSocket Connection Failed

1. **Check if SDK server is running**
   ```bash
   # Test with HTTP first
   curl http://localhost:8971/api/v1/ping
   ```

2. **Check WebSocket endpoint**
   ```javascript
   // Should be accessible at:
   ws://localhost:8971/api/v1/ws/connection
   ```

3. **Check browser console for errors**
   ```javascript
   // Enable debug logging
   sdk.onHealthCheck = (data) => {
       console.log('Health check:', data);
   };
   ```

### CORS Issues

WebSocket doesn't have CORS restrictions, but HTTP API calls do. Make sure server has CORS enabled:

```go
// Server should have:
app.Use(cors.New(cors.Config{
    AllowOrigins: "*",
}))
```

## 📝 API Reference

### Constructor

```javascript
new PrinterSDK(baseUrl?: string)
```
- `baseUrl` (optional): Base URL of the SDK server. Default: `http://localhost:8971/api/v1`

### Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `checkConnection()` | Check if server is accessible | `Promise<boolean>` |
| `startHealthCheck()` | Start WebSocket health monitoring | `void` |
| `stopHealthCheck()` | Stop WebSocket and disconnect | `void` |
| `sendPing()` | Send ping to server | `boolean` |
| `getConnectionStatus()` | Get current connection info | `Object` |
| `getPrinters()` | Get list of available printers | `Promise<Array>` |
| `getPaper(printerId)` | Get paper sizes for printer | `Promise<Array>` |
| `print(printerId, paperSize, file)` | Print file | `Promise<Object>` |

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `isConnected` | `boolean` | Current connection status |
| `onConnectionChange` | `function` | Callback when connection changes |
| `onHealthCheck` | `function` | Callback for heartbeat messages |
| `maxReconnectAttempts` | `number` | Max reconnection attempts (default: 5) |
| `reconnectDelay` | `number` | Delay between reconnects in ms (default: 3000) |

## 🚀 Demo

Buka file `demo.html` di browser untuk melihat contoh lengkap dengan UI interaktif.

## 📄 License

MIT

## 🤝 Support

Untuk masalah atau pertanyaan, silakan hubungi tim development.
