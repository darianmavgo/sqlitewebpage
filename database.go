package sqldoc

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
)

type TableInfo struct {
	Name     string
	Label    string // display name (from _nav or generated)
	Type     string // "table" or "view"
	RowCount int
	ColCount int
	Hidden   bool
}

type TableData struct {
	Info       TableInfo
	Columns    []string
	Rows       [][]string // all values as strings
	Page       int
	TotalPages int
	PageSize   int
}

type NavOverride struct {
	TableName string
	Label     string
	Position  int
	Hidden    bool
}

func DiscoverTables(db *sql.DB) ([]TableInfo, error) {
	rows, err := db.Query("SELECT name, type FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Type); err != nil {
			return nil, err
		}
		t.Label = t.Name
		if strings.HasPrefix(t.Name, "_") {
			t.Hidden = true
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range tables {
		// Row count
		rowQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tables[i].Name)
		_ = db.QueryRow(rowQuery).Scan(&tables[i].RowCount)

		// Column count
		pragmaQuery := fmt.Sprintf(`PRAGMA table_info("%s")`, tables[i].Name)
		pRows, err := db.Query(pragmaQuery)
		if err == nil {
			count := 0
			for pRows.Next() {
				count++
			}
			pRows.Close()
			tables[i].ColCount = count
		}
	}

	return tables, nil
}

func LoadNavOverrides(db *sql.DB) ([]NavOverride, error) {
	var exists string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='_nav'").Scan(&exists)
	if err != nil || exists == "" {
		return nil, nil // Not found or error
	}

	rows, err := db.Query("SELECT table_name, label, position, hidden FROM _nav")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []NavOverride
	for rows.Next() {
		var o NavOverride
		if err := rows.Scan(&o.TableName, &o.Label, &o.Position, &o.Hidden); err != nil {
			return nil, err
		}
		overrides = append(overrides, o)
	}
	return overrides, rows.Err()
}

func PrioritizeTables(tables []TableInfo, overrides []NavOverride) []TableInfo {
	overrideMap := make(map[string]NavOverride)
	for _, o := range overrides {
		overrideMap[o.TableName] = o
	}

	var visibleTables []TableInfo

	for _, t := range tables {
		if o, ok := overrideMap[t.Name]; ok {
			if o.Label != "" {
				t.Label = o.Label
			}
			t.Hidden = o.Hidden
		}
		if !t.Hidden {
			visibleTables = append(visibleTables, t)
		}
	}

	sort.SliceStable(visibleTables, func(i, j int) bool {
		ti := visibleTables[i]
		tj := visibleTables[j]

		oi, hasOi := overrideMap[ti.Name]
		oj, hasOj := overrideMap[tj.Name]

		// 1. Explicit position
		if hasOi && hasOj && oi.Position != oj.Position {
			return oi.Position < oj.Position
		}
		if hasOi && oi.Position > 0 && (!hasOj || oj.Position == 0) {
			return true
		}
		if hasOj && oj.Position > 0 && (!hasOi || oi.Position == 0) {
			return false
		}

		// 2. 'home' in name
		homeI := strings.Contains(strings.ToLower(ti.Name), "home")
		homeJ := strings.Contains(strings.ToLower(tj.Name), "home")
		if homeI != homeJ {
			return homeI
		}

		// 3. Row count > 10
		if ti.RowCount > 10 || tj.RowCount > 10 {
			if ti.RowCount != tj.RowCount {
				return ti.RowCount > tj.RowCount
			}
		}

		// 4. Alphabetically
		return ti.Name < tj.Name
	})

	return visibleTables
}

func FetchTableData(db *sql.DB, tableName string, page int, pageSize int) (*TableData, error) {
	var rowCount int
	rowQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
	if err := db.QueryRow(rowQuery).Scan(&rowCount); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(rowCount) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`SELECT * FROM "%s" LIMIT ? OFFSET ?`, tableName)
	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var resultRows [][]string
	colCount := len(cols)
	values := make([]interface{}, colCount)
	scanArgs := make([]interface{}, colCount)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		strRow := make([]string, colCount)
		for i, v := range values {
			if v == nil {
				strRow[i] = ""
			} else if b, ok := v.([]byte); ok {
				strRow[i] = string(b)
			} else {
				strRow[i] = fmt.Sprintf("%v", v)
			}
		}
		resultRows = append(resultRows, strRow)
	}

	return &TableData{
		Info: TableInfo{
			Name:     tableName,
			RowCount: rowCount,
			ColCount: colCount,
		},
		Columns:    cols,
		Rows:       resultRows,
		Page:       page,
		TotalPages: totalPages,
		PageSize:   pageSize,
	}, nil
}
