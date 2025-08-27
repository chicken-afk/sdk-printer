class PrinterSDK {
    constructor(baseUrl) {
        this.baseUrl = baseUrl || "http://localhost:8971/api/v1";
        this.healthCheckInterval = null;

        // Immediately start health check
        this.startHealthCheck();
    }

    // 🔹 Check if the server is accessible
    async checkConnection() {
        try {
            const controller = new AbortController();
            const timeout = setTimeout(() => controller.abort(), 3000); // 3 seconds timeout

            const res = await fetch(`${this.baseUrl}/printers`, {
                signal: controller.signal,
            });
            console.log(res)
            clearTimeout(timeout);

            if (!res.ok) throw new Error("Health check failed");
            return true;
        } catch (err) {
            return false;
        }
    }

    // 🔹 Run health check every 10 seconds
    startHealthCheck() {
        if (this.healthCheckInterval) return; // prevent duplicate intervals

        this.healthCheckInterval = setInterval(async () => {
            const ok = await this.checkConnection();
            if (!ok) {
                alert("⚠️ Printer SDK is not running. Please open the application.");
                clearInterval(this.healthCheckInterval); // stop checking to avoid alert spam
                this.healthCheckInterval = null;
            }
        }, 10000); // 10 seconds
    }

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
