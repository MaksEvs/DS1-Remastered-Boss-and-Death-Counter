//go:build windows

package main

import (
	"log"
	"net/http"

	webview "github.com/webview/webview_go"
)

func runDesktopWindow(server *http.Server) {
	window := webview.New(false)
	if window == nil {
		log.Fatal("could not create application window")
	}
	defer window.Destroy()

	window.SetTitle("DS1 Remastered — Boss Counter")
	window.SetSize(520, 360, webview.HintNone)
	window.Navigate("http://" + server.Addr)
	window.Run()
	stopHTTPServer(server)
}

func stopHTTPServer(server *http.Server) {
	if err := server.Close(); err != nil {
		log.Printf("HTTP server close failed: %v", err)
	}
}
