class PrinterSDK {
    constructor(baseUrl) {
        this.baseUrl = baseUrl || "http://localhost:8971/api/v1";
        this.healthCheckInterval = null;

        // langsung mulai health check
        this.startHealthCheck();
    }

    // 🔹 Cek apakah server bisa diakses
    async checkConnection() {
        try {
            const controller = new AbortController();
            const timeout = setTimeout(() => controller.abort(), 3000); // timeout 3 detik

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

    // 🔹 Jalanin health check setiap 10 detik
    startHealthCheck() {
        if (this.healthCheckInterval) return; // biar tidak dobel

        this.healthCheckInterval = setInterval(async () => {
            const ok = await this.checkConnection();
            if (!ok) {
                alert("⚠️ Printer SDK belum dijalankan. Harap buka aplikasinya.");
                clearInterval(this.healthCheckInterval); // stop cek biar ga spam alert
                this.healthCheckInterval = null;
            }
        }, 10000); // 10 detik
    }

    // 🔹 Ambil daftar printer
    async getPrinters() {
        const res = await fetch(`${this.baseUrl}/printers`);
        if (!res.ok) throw new Error("Failed to fetch printers");
        return await res.json();
    }

    // 🔹 Ambil daftar kertas dari printer tertentu
    async getPaper(printerId) {
        const res = await fetch(`${this.baseUrl}/printers/${printerId}/papers`);
        if (!res.ok) throw new Error("Failed to fetch paper list");
        return await res.json();
    }

    // 🔹 Kirim file ke printer
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

// Export ke global biar bisa dipakai di browser
window.PrinterSDK = PrinterSDK;
