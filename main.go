package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/kbinani/screenshot"
	"github.com/skip2/go-qrcode"
	webview "github.com/webview/webview_go"

	"screenmirror/stream"
	"screenmirror/utils"
)

//go:embed mobile/* ui/*
var content embed.FS

func main() {
	port := "8080"
	ip, err := utils.GetOutboundIP()
	if err != nil {
		ip = "127.0.0.1"
	}
	serverUrl := fmt.Sprintf("http://%s:%s/mobile/", ip, port)

	videoStreamer := stream.NewStreamer()

	mux := http.NewServeMux()

	mobileFS, _ := fs.Sub(content, "mobile")
	uiFS, _ := fs.Sub(content, "ui")

	mux.Handle("/", http.FileServer(http.FS(mobileFS)))
	mux.Handle("/mobile/", http.StripPrefix("/mobile/", http.FileServer(http.FS(mobileFS))))

	mux.Handle("/desktop", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		indexData, err := content.ReadFile("ui/index.html")
		if err != nil {
			http.Error(w, "Not found", 404)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexData)
	}))
	mux.Handle("/desktop_assets/", http.StripPrefix("/desktop_assets/", http.FileServer(http.FS(uiFS))))

	mux.Handle("/stream", videoStreamer)

	mux.HandleFunc("/api/qr", func(w http.ResponseWriter, r *http.Request) {
		png, err := qrcode.Encode(serverUrl, qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "QR generation failed", 500)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	})

	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"url": serverUrl,
		})
	})

	mux.HandleFunc("/api/displays", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		n := screenshot.NumActiveDisplays()
		json.NewEncoder(w).Encode(map[string]int{
			"count":   n,
			"current": videoStreamer.GetDisplay(),
		})
	})

	mux.HandleFunc("/api/set_display", func(w http.ResponseWriter, r *http.Request) {
		idxStr := r.URL.Query().Get("id")
		var idx int
		fmt.Sscanf(idxStr, "%d", &idx)
		videoStreamer.SetDisplay(idx)
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		localAddr := fmt.Sprintf("0.0.0.0:%s", port)
		log.Printf("Starting server at %s", localAddr)
		if err := http.ListenAndServe(localAddr, mux); err != nil {
			log.Fatal(err)
		}
	}()

	debug := false
	if os.Getenv("DEBUG") == "1" {
		debug = true
	}
	
	w := webview.New(debug)
	defer w.Destroy()

	w.SetTitle("Трансляция Экрана")
	w.SetSize(480, 650, webview.HintNone)
	w.Navigate(fmt.Sprintf("http://127.0.0.1:%s/desktop", port))
	w.Run()
}
