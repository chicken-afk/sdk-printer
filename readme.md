# 🖨️ Printer SDK

Printer SDK is an application that bridges **Web Applications** with **local Printers** via API (HTTP + WebSocket) running in the background (system tray).  
This SDK allows web applications to call functions:
- `getPrinters()`
- `getPaper(printerId)`
- `print(printerId, paperSize, file)`

---

## 🚀 Features
- API Server based on **Go Fiber**.
- **System Tray App** (Windows) with a printer icon.
- **CDN JavaScript SDK** for easy integration with HTML/Frontend.

---

## 📦 Installation

### 1. Build Printer SDK (Go API)
Make sure you have Go (1.20+) installed.

```bash
git clone https://github.com/chicken-afk/printer-sdk.git
cd printer-sdk
go mod tidy
go build -ldflags "-H=windowsgui" -o PrinterSDK.exe
```

Or simply you can download PrinterSDK.exe in builds folder in this repository if you dont want to build it manually, Its on builds/PrinterSDK.exe.

### 2. Running the SDK
Simply double-click the built printer-sdk.exe file.
The application will:

Display an icon in the system tray.

Run the API server at http://localhost:8971.

### 3. Use the Javascript SDK
To use in HTML, refer to the example in printer.html

```html
<body>
    <select id="printerSelect"></select>
    <select id="paperSize"></select>
    <input type="file" id="fileInput" />
    <button onclick="testPrint()">Test Print</button>

    <script src="https://cdn.jsdelivr.net/gh/chicken-afk/sdk-printer@latest/cdn/printer-sdk.js"></script>
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
```