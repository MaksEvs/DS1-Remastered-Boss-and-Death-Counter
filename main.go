package main

import (
	"embed"
	"log"
	"net/http"
	"os"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--diagnostic" {
		diagnosticMain()
		return
	}

	applicationState := newApplicationState()
	loadApplicationState(applicationState)
	startGameMonitor(applicationState)

	handler := newHTTPHandler(applicationState, webFiles)
	server := &http.Server{Addr: "127.0.0.1:7373", Handler: handler}
	log.Printf("Application: http://%s", server.Addr)
	log.Printf("OBS text: http://%s/obs.txt", server.Addr)

	go startHTTPServer(server)
	runDesktopWindow(server)
}

func startHTTPServer(server *http.Server) {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server stopped: %v", err)
	}
}
