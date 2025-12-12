class PrinterSDK {
    constructor(baseUrl) {
        this.baseUrl = baseUrl || "http://localhost:8971/api/v1";
        this.wsUrl = (baseUrl || "http://localhost:8971/api/v1").replace(/^http/, 'ws') + "/ws/connection";
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 100;
        this.reconnectDelay = 3000; // 3 seconds
        this.onConnectionChange = null; // callback untuk notifikasi perubahan koneksi
        this.onHealthCheck = null; // callback untuk heartbeat
        this.connectionTimeout = null;

        // Printer monitoring WebSocket
        this.printerWs = null;
        this.printerMonitoringActive = false; // flag untuk tracking apakah monitoring aktif
        this.onPrinterStatusUpdate = null; // callback untuk printer status update
        this.onPrinterMonitorConnected = null; // callback ketika printer monitor connected

        // Immediately start health check via WebSocket
        this.startHealthCheck();
    }

    // 🔹 Check if the server is accessible via WebSocket
    async checkConnection() {
        return new Promise((resolve) => {
            if (this.isConnected) {
                resolve(true);
                return;
            }

            const testWs = new WebSocket(this.wsUrl);
            const timeout = setTimeout(() => {
                testWs.close();
                resolve(false);
            }, 3000);

            testWs.onopen = () => {
                clearTimeout(timeout);
                testWs.close();
                resolve(true);
            };

            testWs.onerror = () => {
                clearTimeout(timeout);
                resolve(false);
            };
        });
    }

    // 🔹 Start WebSocket connection for real-time health monitoring
    startHealthCheck() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            console.log("WebSocket already connected");
            return;
        }

        console.log("Connecting to WebSocket...", this.wsUrl);
        this.ws = new WebSocket(this.wsUrl);

        // Set connection timeout
        this.connectionTimeout = setTimeout(() => {
            if (!this.isConnected) {
                console.error("WebSocket connection timeout");
                this.ws.close();
                this.handleDisconnect();
            }
        }, 5000);

        this.ws.onopen = () => {
            clearTimeout(this.connectionTimeout);
            this.isConnected = true;
            this.reconnectAttempts = 0;
            console.log("✅ WebSocket connected to Printer SDK");

            // Notify connection change
            if (this.onConnectionChange) {
                this.onConnectionChange(true);
            }
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                console.log("📨 Health check:", data);

                // Notify health check callback
                if (this.onHealthCheck) {
                    this.onHealthCheck(data);
                }

                // Update connection status
                if (data.status === "online" || data.status === "connected") {
                    this.isConnected = true;
                }
            } catch (err) {
                console.error("Error parsing WebSocket message:", err);
            }
        };

        this.ws.onerror = (error) => {
            console.error("❌ WebSocket error:", error);
        };

        this.ws.onclose = () => {
            clearTimeout(this.connectionTimeout);
            console.log("WebSocket closed");
            this.handleDisconnect();
        };
    }

    // 🔹 Handle disconnect and attempt reconnection
    handleDisconnect() {
        this.isConnected = false;
        this.ws = null;

        // Notify connection change
        if (this.onConnectionChange) {
            this.onConnectionChange(false);
        }

        // Attempt to reconnect
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

            setTimeout(() => {
                this.startHealthCheck();
            }, this.reconnectDelay);
        } else {
            console.error("⚠️ Max reconnection attempts reached. Please check if Printer SDK is running.");
            if (this.onConnectionChange) {
                this.onConnectionChange(false, true); // true = max retries reached
            }
        }
    }

    // 🔹 Manually send ping to server
    sendPing() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'ping',
                timestamp: new Date().toISOString()
            }));
            return true;
        }
        return false;
    }

    // 🔹 Get current connection status
    getConnectionStatus() {
        return {
            isConnected: this.isConnected,
            reconnectAttempts: this.reconnectAttempts,
            wsReadyState: this.ws ? this.ws.readyState : null
        };
    }

    // 🔹 Stop health check and close WebSocket
    stopHealthCheck() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.isConnected = false;
        if (this.connectionTimeout) {
            clearTimeout(this.connectionTimeout);
        }
    }

    // ========== PRINTER MONITORING WebSocket ==========

    // 🔹 Start monitoring printer status in real-time
    startPrinterMonitoring() {
        // Cleanup existing connection first to prevent duplicate connections
        if (this.printerWs) {
            if (this.printerWs.readyState === WebSocket.OPEN) {
                console.log("Printer monitoring WebSocket already connected");
                return;
            }
            // Close existing connection that might be in CONNECTING or CLOSING state
            try {
                this.printerWs.close();
            } catch (e) {
                console.log("Error closing old connection:", e);
            }
            this.printerWs = null;
        }

        this.printerMonitoringActive = true; // Set flag bahwa monitoring aktif

        const printerWsUrl = this.wsUrl.replace('/ws/connection', '/ws/printers');
        console.log("Connecting to Printer Monitor WebSocket...", printerWsUrl);

        this.printerWs = new WebSocket(printerWsUrl);

        this.printerWs.onopen = () => {
            console.log("✅ Printer monitoring WebSocket connected");
            if (this.onPrinterMonitorConnected) {
                this.onPrinterMonitorConnected();
            }
        };

        this.printerWs.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                console.log("🖨️ Printer status update:", data);

                // Notify printer status callback
                if (this.onPrinterStatusUpdate && data.type === "printer_status") {
                    this.onPrinterStatusUpdate(data);
                }
            } catch (err) {
                console.error("Error parsing printer monitor message:", err);
            }
        };

        this.printerWs.onerror = (error) => {
            console.error("❌ Printer monitoring WebSocket error:", error);
        };

        this.printerWs.onclose = () => {
            console.log("Printer monitoring WebSocket closed");
            this.printerWs = null;

            // Auto reconnect HANYA jika monitoring masih aktif dan SDK terkoneksi
            if (this.printerMonitoringActive && this.isConnected) {
                console.log("Auto-reconnecting printer monitor...");
                setTimeout(() => {
                    this.startPrinterMonitoring();
                }, 3000);
            }
        };
    }

    // 🔹 Request manual refresh of printer status
    refreshPrinterStatus() {
        if (this.printerWs && this.printerWs.readyState === WebSocket.OPEN) {
            this.printerWs.send(JSON.stringify({
                type: 'refresh'
            }));
            return true;
        }
        return false;
    }

    // 🔹 Stop printer monitoring
    stopPrinterMonitoring() {
        console.log("Stopping printer monitoring...");
        this.printerMonitoringActive = false; // Set flag bahwa monitoring tidak aktif

        if (this.printerWs) {
            this.printerWs.close();
            this.printerWs = null;
        }
    }

    // ========== HTTP API Methods ==========

    // 🔹 Get list of printers
    async getPrinters() {
        const res = await fetch(`${this.baseUrl}/printers`);
        if (!res.ok) throw new Error("Failed to fetch printers");
        return await res.json();
    }

    // 🔹 Get list of papers from a specific printer
    async getPaper(printerId) {
        const res = await fetch(`${this.baseUrl}/printers/${printerId}/papers`);
        if (!res.ok) throw new Error("Failed to fetch paper list");
        return await res.json();
    }

    // 🔹 Send file to printer
    async print(printerId, paperSize, file) {
        const formData = new FormData();
        formData.append("printerId", printerId);
        formData.append("paperSize", paperSize);
        formData.append("file", file);

        const res = await fetch(`${this.baseUrl}/print`, {
            method: "POST",
            body: formData,
        });

        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || "Failed to print");
        }

        return await res.json();
    }
}

// Export to global so it can be used in the browser
window.PrinterSDK = PrinterSDK;
