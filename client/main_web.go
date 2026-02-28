package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
)

//go:embed web
var webFS embed.FS

func main() {
	port := "8080"
	if p := os.Getenv("STALKNET_WEB_PORT"); p != "" {
		port = p
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			data, err := webFS.ReadFile("web/index.html")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Write(data)
			return
		}
		if r.URL.Path == "/app.js" {
			data, err := webFS.ReadFile("web/app.js")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/javascript")
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Write(data)
			return
		}
		http.NotFound(w, r)
	})

	addr := ":" + port
	fmt.Printf("STALKnet Web Client starting on http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop")
	log.Fatal(http.ListenAndServe(addr, nil))
}
