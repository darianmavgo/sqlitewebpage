package sqldoc

import (
	"bytes"
	"database/sql"
	"fmt"
)

// RenderDatabase is a high-level convenience function that opens a database,
// loads all metadata, discovers and prioritizes tables, fetches data, and
// renders everything to a self-contained HTML document.
//
// The caller must ensure an appropriate SQLite driver (e.g. modernc.org/sqlite)
// is registered before calling this function.
func RenderDatabase(db *sql.DB) (htmlBytes []byte, head *HeadConfig, tables []TableInfo, err error) {
	// Load head metadata & style
	h := LoadHead(db)

	// Discover and prioritize tables
	allTables, err := DiscoverTables(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("discover tables: %w", err)
	}

	overrides, _ := LoadNavOverrides(db)
	prioritized := PrioritizeTables(allTables, overrides)

	// Fetch all data for each visible table
	var contents []TableData
	for _, t := range prioritized {
		td, fetchErr := FetchTableData(db, t.Name, 1, t.RowCount+1)
		if fetchErr != nil {
			continue
		}
		td.Info.Label = t.Label
		td.Info.Type = t.Type
		td.Info.ColCount = t.ColCount
		td.Info.Hidden = t.Hidden
		td.PageSize = h.Style.PageSize
		td.TotalPages = (t.RowCount + h.Style.PageSize - 1) / h.Style.PageSize
		if td.TotalPages == 0 {
			td.TotalPages = 1
		}
		contents = append(contents, *td)
	}

	data := RenderData{
		Head:     h,
		Style:    h.Style,
		Tables:   prioritized,
		Contents: contents,
	}

	var buf bytes.Buffer
	if err := RenderHTML(data, &buf); err != nil {
		return nil, nil, nil, fmt.Errorf("render HTML: %w", err)
	}

	return buf.Bytes(), &h, prioritized, nil
}
