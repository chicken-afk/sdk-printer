# 🖨️ Printer SDK

Printer SDK adalah aplikasi untuk menjembatani **Web Application** dengan **Printer lokal** melalui API (HTTP + WebSocket) yang berjalan di background (system tray).  
SDK ini memungkinkan aplikasi web memanggil fungsi:
- `getPrinters()`
- `getPaper(printerId)`
- `print(printerId, paperSize, file)`

---

## 🚀 Features
- API Server berbasis **Go Fiber**.
- **WebSocket** untuk monitoring status printer secara realtime.
- **System Tray App** (Windows) dengan ikon printer.
- **CDN JavaScript SDK** untuk integrasi ke HTML/Frontend.

---

## 📦 Installation

### 1. Build Printer SDK (Go API)
Pastikan sudah install Go (1.20+).

```bash
git clone https://github.com/chicken-afk/printer-sdk.git
cd printer-sdk
go mod tidy
go build -ldflags "-H=windowsgui" -o PrinterSDK.exe
```
### 2. Running SDK
Kemudian, cukup double click printer-sdk.exe hasil build sebelumna.
Aplikasi akan:

Menampilkan ikon di system tray.

Menjalankan API server di http://localhost:8971.

### 3. Use Javascript SDK
Lalu cara penggunaan di html adalah seperti dalam file printer.html

```bash
<body>
    <select id="printerSelect"></select>
    <select id="paperSize"></select>
    <input type="file" id="fileInput" />
    <button onclick="testPrint()">Test Print</button>

    <script src="https://cdn.jsdelivr.net/gh/chicken-afk/sdk-printer@latest/printer-sdk.js"></script>
    <script>
        const sdk = new PrinterSDK();

        async function testPrint() {
            const printerSelect = document.getElementById("printerSelect");
            const paperSizeSelect = document.getElementById("paperSize");
            const fileInput = document.querySelector("#fileInput");

            const printerId = printerSelect.value;
            const paper = paperSizeSelect.value;
            const file = fileInput.files[0];

            if (!printerId || !paper || !file) {
                alert("Please select a printer, paper size, and file.");
                return;
            }

            console.log("Selected printer:", printerId);
            console.log("Selected paper:", paper);

            const result = await sdk.print(printerId, paper, file);
            console.log("Print result:", result);
        }

    </script>
    <script>
        // Populate printer select on page load
        document.addEventListener("DOMContentLoaded", async () => {
            const printers = await sdk.getPrinters();
            const printerSelect = document.getElementById("printerSelect");
            const paperSizeSelect = document.getElementById("paperSize");
            printers.forEach(printer => {
                const option = document.createElement("option");
                option.value = printer.id;
                option.textContent = printer.name || printer.id;
                printerSelect.appendChild(option);
            });

            // If printers exist, trigger paper size load for the first printer
            if (printers.length > 0) {
                await updatePaperSizes(printers[0].id);
            }

            // Listen for printer selection change
            printerSelect.addEventListener("change", async (e) => {
                const printerId = e.target.value;
                await updatePaperSizes(printerId);
            });

            async function updatePaperSizes(printerId) {
                // Clear previous options
                paperSizeSelect.innerHTML = "";
                const papers = await sdk.getPaper(printerId);
                papers.forEach(paper => {
                    const option = document.createElement("option");
                    option.value = paper;
                    option.textContent = paper;
                    paperSizeSelect.appendChild(option);
                });
            }
        });
    </script>
</body>
bash```