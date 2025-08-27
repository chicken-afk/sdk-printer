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
git clone https://github.com/<username>/printer-sdk.git
cd printer-sdk
go mod tidy
go build -o printer-sdk.exe
```

