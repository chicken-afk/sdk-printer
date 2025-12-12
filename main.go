package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"os/signal"

	_ "embed"

	"github.com/getlantern/systray"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pos/sdk/router"
)

var app *fiber.App

//go:embed assets/printer.ico
var iconData []byte

func runServer() {
	app = fiber.New()

	// Allow all origins
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	router.Router(app)

	go func() {
		log.Println("Server started at :8971")
		if err := app.Listen(":8971"); err != nil {
			log.Printf("Fiber server closed: %v\n", err)
		}
	}()
}

// ------------ SYSTRAY ------------

func onReady() {
	systray.SetTitle("Printer SDK")
	systray.SetTooltip("Printer SDK Running")
	systray.SetIcon(iconData)

	mOpen := systray.AddMenuItem("Open API Docs", "Open API in browser")
	mQuit := systray.AddMenuItem("Quit", "Stop service and quit")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				// buka API di browser
				fmt.Println("Open clicked: http://localhost:8080/ping")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	fmt.Println("Tray exiting... stopping Fiber server")
	if app != nil {
		_ = app.Shutdown() // <-- ini akan matikan server API
	}
	// Exit program completely when tray is closed
	os.Exit(0)
}

func main() {
	// Jalankan Fiber API di background
	runServer()

	// Handle signal biar bisa shutdown bersih via CTRL+C juga
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Channel untuk notifikasi dari systray
	done := make(chan struct{})

	// Jalankan system tray
	go func() {
		systray.Run(onReady, onExit)
		close(done) // Notify when systray exits
	}()

	// Wait for either signal or systray exit
	select {
	case <-c:
		fmt.Println("Received interrupt signal, shutting down...")
		// Quit systray first to unblock systray.Run()
		systray.Quit()
		// Wait for systray to finish
		<-done
		// Then shutdown Fiber server
		if app != nil {
			_ = app.Shutdown()
		}
		fmt.Println("Server stopped gracefully")
		os.Exit(0)
	case <-done:
		fmt.Println("Systray closed, shutting down...")
		// Systray already closed, just shutdown Fiber
		if app != nil {
			_ = app.Shutdown()
		}
		fmt.Println("Server stopped gracefully")
		os.Exit(0)
	}
}
