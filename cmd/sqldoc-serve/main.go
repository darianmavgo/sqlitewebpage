package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"sqldoc"
	_ "modernc.org/sqlite"
)

func main() {
	log.SetFlags(0)

	var dbPath string
	var port string

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-p" || arg == "--port" {
			if i+1 < len(os.Args) {
				port = os.Args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "-p=") {
			port = strings.TrimPrefix(arg, "-p=")
		} else if !strings.HasPrefix(arg, "-") && dbPath == "" {
			dbPath = arg
		}
	}

	if dbPath == "" {
		fmt.Fprintf(os.Stderr, `sqldoc-serve — serve a SQLite database as an interactive HTML document

Usage:
  sqldoc-serve <database.db> [-p port]

Options:
  -p port   HTTP port (default: auto-select a free port)

The server renders the database to HTML with embedded AG Grid and opens
your default browser. Press Ctrl+C to stop.
`)
		os.Exit(1)
	}

	// Open database and render
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	htmlBytes, head, tables, err := sqldoc.RenderDatabase(db)
	if err != nil {
		log.Fatalf("failed to render: %v", err)
	}

	// Set up HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(htmlBytes)
	})

	// Find a free port if not specified
	listenAddr := "127.0.0.1:" + port
	if port == "" {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("failed to find free port: %v", err)
		}
		listenAddr = listener.Addr().String()
		listener.Close()
	}

	url := "http://" + listenAddr
	title := head.Title
	if title == "" {
		title = dbPath
	}

	log.Printf("📊 %s", title)
	log.Printf("   %d tables rendered from %s", len(tables), dbPath)
	log.Printf("   serving at %s", url)
	log.Printf("   press Ctrl+C to stop")

	// Open browser
	go openBrowser(url)

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("\n✓ shutting down")
		os.Exit(0)
	}()

	// Start server
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		log.Printf("open %s in your browser", url)
		return
	}
	cmd.Start()
}
