package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pos/sdk/handlers"
)

func Router(r *fiber.App) *fiber.App {

	// Handlers printer
	printerHandler := handlers.NewPrinterService()

	// Group routes under /api/v1
	routeV1 := r.Group("/api/v1")

	// Example route: GET /api/v1/ping
	routeV1.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	// Printer routes
	routeV1.Post("/print", printerHandler.Print)
	routeV1.Get("/printers", printerHandler.GetPrinters)
	routeV1.Get("/printers/:id/papers", printerHandler.GetPrinterPapers)

	// No need to use r.Use(routeV1), group already registered
	return r
}
