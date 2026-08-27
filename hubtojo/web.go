package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

//go:embed templates/stats.html
var templateFS embed.FS

var statsPage = template.Must(template.New("stats").Funcs(template.FuncMap{
	"formatTime":     formatWebTime,
	"formatDuration": formatWebDuration,
}).ParseFS(templateFS, "templates/stats.html"))

func StartWebServer(addr string, store *StatsStore, metrics *Metrics) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := statsPage.ExecuteTemplate(w, "stats.html", store.Snapshot()); err != nil {
			log.Printf("Error rendering stats page: %s\n", err)
		}
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(store.Snapshot()); err != nil {
			log.Printf("Error encoding stats: %s\n", err)
		}
	})
	mux.Handle("/metrics", metrics.Handler())

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server.Addr = listener.Addr().String()
	go func() {
		log.Printf("Web server listening on %s\n", listener.Addr())
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Web server error: %s\n", err)
		}
	}()
	return server, nil
}

func formatWebTime(value any) string {
	switch t := value.(type) {
	case time.Time:
		if t.IsZero() {
			return "unknown"
		}
		return t.Format(time.RFC3339)
	case *time.Time:
		if t == nil || t.IsZero() {
			return "unknown"
		}
		return t.Format(time.RFC3339)
	default:
		return "unknown"
	}
}

func formatWebDuration(seconds float64) string {
	return (time.Duration(seconds * float64(time.Second))).Round(time.Millisecond).String()
}
