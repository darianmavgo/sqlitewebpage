package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"sqldoc"
	_ "modernc.org/sqlite"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "render":
		renderCmd(os.Args[2:])
	case "info":
		infoCmd(os.Args[2:])
	default:
		log.Printf("unknown command: %s", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `sqldoc — render SQLite databases as styled HTML documents

Usage:
  sqldoc render <database.db> [-o output.html]
  sqldoc info   <database.db>

Commands:
  render    Render any SQLite database (.db, .sqlite, .sqlite3) to a self-contained HTML file
  info      Show table summary and priority order for a database
`)
}

func renderCmd(args []string) {
	var output string
	var dbPath string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-o" || arg == "--output" {
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "-o=") {
			output = strings.TrimPrefix(arg, "-o=")
		} else if strings.HasPrefix(arg, "--output=") {
			output = strings.TrimPrefix(arg, "--output=")
		} else if !strings.HasPrefix(arg, "-") && dbPath == "" {
			dbPath = arg
		}
	}

	if dbPath == "" {
		log.Fatal("usage: sqldoc render <database.db> [-o output.html]")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	htmlBytes, _, tables, err := sqldoc.RenderDatabase(db)
	if err != nil {
		log.Fatalf("failed to render: %v", err)
	}

	if output != "" {
		if err := os.WriteFile(output, htmlBytes, 0644); err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
		log.Printf("✓ rendered %d tables → %s", len(tables), output)
	} else {
		os.Stdout.Write(htmlBytes)
	}
}

func infoCmd(args []string) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		log.Fatal("usage: sqldoc info <database.db>")
	}
	dbPath := fs.Arg(0)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	head := sqldoc.LoadHead(db)
	fmt.Printf("Document: %s\n", head.Title)
	if head.Description != "" {
		fmt.Printf("Description: %s\n", head.Description)
	}
	if head.Author != "" {
		fmt.Printf("Author: %s\n", head.Author)
	}
	if head.Style.AccentColor != "" {
		fmt.Printf("Theme: Accent=%s, DarkMode=%s, PageSize=%d\n", head.Style.AccentColor, head.Style.DarkMode, head.Style.PageSize)
	}
	fmt.Println()

	allTables, err := sqldoc.DiscoverTables(db)
	if err != nil {
		log.Fatalf("failed to discover tables: %v", err)
	}

	overrides, err := sqldoc.LoadNavOverrides(db)
	if err != nil {
		log.Fatalf("failed to load nav overrides: %v", err)
	}

	tables := sqldoc.PrioritizeTables(allTables, overrides)

	fmt.Printf("%-4s  %-25s  %-25s  %-6s  %5s  %4s\n", "#", "Table", "Label", "Type", "Rows", "Cols")
	fmt.Println("----  -------------------------  -------------------------  ------  -----  ----")
	for i, t := range tables {
		fmt.Printf("%-4d  %-25s  %-25s  %-6s  %5d  %4d\n",
			i+1, t.Name, t.Label, t.Type, t.RowCount, t.ColCount)
	}

	fmt.Println("\nHidden/Metadata tables:")
	for _, t := range allTables {
		if t.Hidden {
			fmt.Printf("  %-25s  %-6s  %5d rows\n", t.Name, t.Type, t.RowCount)
		}
	}
}
