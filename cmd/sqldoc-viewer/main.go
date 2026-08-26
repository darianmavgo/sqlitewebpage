package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"sqldoc"
	webview "github.com/webview/webview_go"
	_ "modernc.org/sqlite"
)

func main() {
	log.SetFlags(0)

	var dbPath string

	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		dbPath = os.Args[1]
	}

	// If no database specified and running on macOS, prompt with native file picker
	if dbPath == "" && runtime.GOOS == "darwin" {
		picked, err := pickFileMacOS()
		if err == nil && picked != "" {
			dbPath = picked
		}
	}

	if dbPath == "" {
		fmt.Fprintf(os.Stderr, `sqldoc-viewer — open a SQLite database as an interactive document in a native window

Usage:
  sqldoc-viewer <database.db>

Opens a native desktop window (using the system WebKit engine) displaying
the database as an interactive HTML document with AG Grid. No browser or
network connection required.
`)
		os.Exit(1)
	}

	// Open database and render
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	htmlBytes, head, tables, err := sqldoc.RenderDatabase(db)
	db.Close() // Close DB before entering GUI loop
	if err != nil {
		log.Fatalf("failed to render: %v", err)
	}

	title := head.Title
	if title == "" {
		title = dbPath
	}

	log.Printf("📊 %s — %d tables from %s", title, len(tables), dbPath)

	// Create native webview window
	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle(title)
	w.SetSize(1280, 800, webview.HintNone)
	w.SetHtml(string(htmlBytes))

	w.Run()
}

func pickFileMacOS() (string, error) {
	script := `POSIX path of (choose file with prompt "Select a SQLite Database:" of type {"public.data", "db", "sqlite", "sqlite3"})`
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
