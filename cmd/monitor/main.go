package main

import (
	"log"

	"github.com/carlosj934/laptop-dashboard-alerting/internal/tray"
)

func main() {
	log.Println("Starting Laptop Monitor...")

	app, err := tray.NewApp()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	log.Println("Laptop Monitor running in system tray")
	app.Run()
}
