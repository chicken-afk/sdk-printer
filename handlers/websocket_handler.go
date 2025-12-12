package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pos/sdk/utils"
)

type ConnectionStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// WebSocket handler untuk connection check
func HandleWebSocket(c *websocket.Conn) {
	defer c.Close()

	log.Printf("WebSocket client connected: %s", c.RemoteAddr())

	// Send initial connection message
	initialMsg := ConnectionStatus{
		Status:    "connected",
		Timestamp: time.Now(),
		Message:   "SDK Printer Server is running",
	}
	if err := c.WriteJSON(initialMsg); err != nil {
		log.Printf("Error sending initial message: %v", err)
		return
	}

	// Ticker untuk mengirim ping setiap 30 detik
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Channel untuk handle graceful shutdown
	done := make(chan struct{})

	// Goroutine untuk membaca message dari client (untuk detect disconnect)
	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}

			// Handle ping dari client
			if messageType == websocket.TextMessage {
				var msg map[string]interface{}
				if err := json.Unmarshal(message, &msg); err == nil {
					if msg["type"] == "ping" {
						// Respond dengan pong
						pongMsg := ConnectionStatus{
							Status:    "online",
							Timestamp: time.Now(),
							Message:   "pong",
						}
						if err := c.WriteJSON(pongMsg); err != nil {
							log.Printf("Error sending pong: %v", err)
							return
						}
					}
				}
			}
		}
	}()

	// Loop untuk mengirim heartbeat
	for {
		select {
		case <-done:
			log.Printf("WebSocket client disconnected: %s", c.RemoteAddr())
			return
		case <-ticker.C:
			// Kirim heartbeat setiap 30 detik
			heartbeat := ConnectionStatus{
				Status:    "online",
				Timestamp: time.Now(),
				Message:   "heartbeat",
			}
			if err := c.WriteJSON(heartbeat); err != nil {
				log.Printf("Error sending heartbeat: %v", err)
				return
			}
		}
	}
}

// Middleware untuk upgrade HTTP ke WebSocket
func WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// PrinterStatus represents the status of a printer
type PrinterStatus struct {
	Type       string                `json:"type"`
	Timestamp  time.Time             `json:"timestamp"`
	Printers   []PrinterStatusDetail `json:"printers"`
	TotalCount int                   `json:"total_count"`
}

type PrinterStatusDetail struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"` // "online", "offline", "unknown"
	PaperSizes []string `json:"paper_sizes"`
}

// HandlePrinterMonitor - WebSocket handler untuk monitoring printer status
func HandlePrinterMonitor(c *websocket.Conn) {
	defer c.Close()

	log.Printf("Printer Monitor WebSocket client connected: %s", c.RemoteAddr())

	// Send initial printer list
	sendPrinterStatus(c)

	// Ticker untuk update status setiap 5 detik
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Channel untuk handle graceful shutdown
	done := make(chan struct{})

	// Goroutine untuk membaca message dari client
	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Printer Monitor WebSocket read error: %v", err)
				}
				return
			}

			// Handle request dari client
			if messageType == websocket.TextMessage {
				var msg map[string]interface{}
				if err := json.Unmarshal(message, &msg); err == nil {
					if msg["type"] == "refresh" {
						// Client request refresh printer list
						sendPrinterStatus(c)
					}
				}
			}
		}
	}()

	// Loop untuk mengirim status update
	for {
		select {
		case <-done:
			log.Printf("Printer Monitor WebSocket client disconnected: %s", c.RemoteAddr())
			return
		case <-ticker.C:
			// Kirim status update setiap 5 detik
			sendPrinterStatus(c)
		}
	}
}

// sendPrinterStatus sends current printer status to WebSocket client
func sendPrinterStatus(c *websocket.Conn) {
	printers := utils.GetPrinterList()

	var printerDetails []PrinterStatusDetail
	for _, p := range printers {
		// Get real printer status from Windows
		status := utils.GetPrinterStatus(p.Name)

		printerDetails = append(printerDetails, PrinterStatusDetail{
			ID:         p.ID,
			Name:       p.Name,
			Status:     status,
			PaperSizes: p.PaperSizes,
		})
	}

	statusMsg := PrinterStatus{
		Type:       "printer_status",
		Timestamp:  time.Now(),
		Printers:   printerDetails,
		TotalCount: len(printerDetails),
	}

	if err := c.WriteJSON(statusMsg); err != nil {
		log.Printf("Error sending printer status: %v", err)
	}
}
