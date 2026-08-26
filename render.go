package sqldoc

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
)

// Embed minified AG Grid Community CSS and JS
//
//go:embed assets/ag-grid.min.css
var agGridCSS string

//go:embed assets/ag-theme-quartz.min.css
var agThemeQuartzCSS string

//go:embed assets/ag-grid-community.min.js
var agGridJS string

type RenderData struct {
	Head     HeadConfig
	Style    StyleConfig
	Tables   []TableInfo
	Contents []TableData
}

func RenderHTML(data RenderData, w io.Writer) error {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"seq": func(n int) []int {
			seq := make([]int, n)
			for i := range seq {
				seq[i] = i + 1
			}
			return seq
		},
		"slugify": func(s string) string {
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, " ", "-")
			s = strings.ReplaceAll(s, "_", "-")
			return s
		},
		"gt": func(a, b int) bool { return a > b },
		"pageForRow": func(idx, pageSize int) int {
			if pageSize <= 0 {
				return 1
			}
			return (idx / pageSize) + 1
		},
		"toGridJSON": func(td TableData) (template.JS, error) {
			var rows []map[string]interface{}
			for _, row := range td.Rows {
				obj := make(map[string]interface{}, len(td.Columns))
				for i, col := range td.Columns {
					if i < len(row) {
						obj[col] = row[i]
					} else {
						obj[col] = ""
					}
				}
				rows = append(rows, obj)
			}
			payload := map[string]interface{}{
				"tableName": td.Info.Name,
				"label":     td.Info.Label,
				"columns":   td.Columns,
				"rows":      rows,
				"colCount":  td.Info.ColCount,
				"rowCount":  td.Info.RowCount,
				"pageSize":  td.PageSize,
			}
			b, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			return template.JS(b), nil
		},
		"embedAGGridCSS": func() template.CSS {
			return template.CSS(agGridCSS + "\n" + agThemeQuartzCSS)
		},
		"embedAGGridJS": func() template.JS {
			return template.JS(agGridJS)
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"renderScript": func(s string) template.HTML {
			trimmed := strings.TrimSpace(s)
			if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "./") {
				return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, template.HTMLEscapeString(trimmed)))
			}
			return template.HTML(fmt.Sprintf(`<script>%s</script>`, trimmed))
		},
	}

	tmpl, err := template.New("index").Funcs(funcs).Parse(htmlTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en" class="{{if eq .Style.DarkMode "dark"}}dark{{end}}">
<head>
<meta charset="{{if .Head.Charset}}{{.Head.Charset}}{{else}}UTF-8{{end}}">
<meta name="viewport" content="{{if .Head.Viewport}}{{.Head.Viewport}}{{else}}width=device-width, initial-scale=1.0{{end}}">

<title>{{if .Head.Title}}{{.Head.Title}}{{else if .Style.Title}}{{.Style.Title}}{{else}}SQLite Document{{end}}</title>

{{if .Head.BaseHref}}
<base href="{{.Head.BaseHref}}" {{if .Head.BaseTarget}}target="{{.Head.BaseTarget}}"{{end}}>
{{end}}

<!-- Standard Meta Tags -->
{{if .Head.Description}}<meta name="description" content="{{.Head.Description}}">{{end}}
{{if .Head.Keywords}}<meta name="keywords" content="{{.Head.Keywords}}">{{end}}
{{if .Head.Author}}<meta name="author" content="{{.Head.Author}}">{{end}}
{{if .Head.Robots}}<meta name="robots" content="{{.Head.Robots}}">{{end}}
{{if .Head.Generator}}<meta name="generator" content="{{.Head.Generator}}">{{end}}
{{if .Head.ThemeColor}}<meta name="theme-color" content="{{.Head.ThemeColor}}">{{end}}
{{if .Head.ColorScheme}}<meta name="color-scheme" content="{{.Head.ColorScheme}}">{{end}}

<!-- OpenGraph Meta Tags -->
{{if .Head.OGTitle}}<meta property="og:title" content="{{.Head.OGTitle}}">{{end}}
{{if .Head.OGDescription}}<meta property="og:description" content="{{.Head.OGDescription}}">{{end}}
{{if .Head.OGImage}}<meta property="og:image" content="{{.Head.OGImage}}">{{end}}
{{if .Head.OGUrl}}<meta property="og:url" content="{{.Head.OGUrl}}">{{end}}
{{if .Head.OGType}}<meta property="og:type" content="{{.Head.OGType}}">{{end}}
{{if .Head.OGSiteName}}<meta property="og:site_name" content="{{.Head.OGSiteName}}">{{end}}

<!-- Twitter Meta Tags -->
{{if .Head.TwitterCard}}<meta name="twitter:card" content="{{.Head.TwitterCard}}">{{end}}
{{if .Head.TwitterTitle}}<meta name="twitter:title" content="{{.Head.TwitterTitle}}">{{end}}
{{if .Head.TwitterDescription}}<meta name="twitter:description" content="{{.Head.TwitterDescription}}">{{end}}
{{if .Head.TwitterImage}}<meta name="twitter:image" content="{{.Head.TwitterImage}}">{{end}}
{{if .Head.TwitterSite}}<meta name="twitter:site" content="{{.Head.TwitterSite}}">{{end}}
{{if .Head.TwitterCreator}}<meta name="twitter:creator" content="{{.Head.TwitterCreator}}">{{end}}

<!-- Custom Meta Tags -->
{{range .Head.CustomMetas}}
<meta {{if .Name}}name="{{.Name}}"{{end}} {{if .Property}}property="{{.Property}}"{{end}} {{if .HttpEquiv}}http-equiv="{{.HttpEquiv}}"{{end}} content="{{.Content}}">
{{end}}

<!-- Link Tags & Icons -->
{{if .Head.Favicon}}<link rel="icon" href="{{safeURL .Head.Favicon}}">{{end}}
{{if .Head.AppleTouchIcon}}<link rel="apple-touch-icon" href="{{safeURL .Head.AppleTouchIcon}}">{{end}}
{{if .Head.Canonical}}<link rel="canonical" href="{{safeURL .Head.Canonical}}">{{end}}

{{range .Head.Preconnects}}
<link rel="preconnect" href="{{safeURL .}}">
{{end}}

{{range .Head.Stylesheets}}
<link rel="stylesheet" href="{{safeURL .}}">
{{end}}

{{range .Head.CustomLinks}}
<link rel="{{.Rel}}" href="{{safeURL .Href}}" {{if .Type}}type="{{.Type}}"{{end}} {{if .As}}as="{{.As}}"{{end}} {{if .CrossOrigin}}crossorigin="{{.CrossOrigin}}"{{end}}>
{{end}}

<!-- Custom Head Scripts -->
{{range .Head.HeadScripts}}
{{renderScript .}}
{{end}}

<!-- Raw Head Content -->
{{if .Head.RawHead}}
{{safeHTML .Head.RawHead}}
{{end}}

<!-- Embedded AG Grid CSS -->
<style>
{{embedAGGridCSS}}
</style>

<!-- Application & Style CSS -->
<style>
:root {
	--accent: {{if .Style.AccentColor}}{{.Style.AccentColor}}{{else}}#2563eb{{end}};
	--accent-hover: #1d4ed8;
	--bg: {{if .Style.BgColor}}{{.Style.BgColor}}{{else}}#ffffff{{end}};
	--text: {{if .Style.TextColor}}{{.Style.TextColor}}{{else}}#1e293b{{end}};
	--font: {{if .Style.FontFamily}}{{.Style.FontFamily}}{{else}}system-ui, -apple-system, sans-serif{{end}};
	
	--sidebar-bg: {{if .Style.SidebarBg}}{{.Style.SidebarBg}}{{else}}#f8fafc{{end}};
	--sidebar-border: #e2e8f0;
	--border: #e2e8f0;
	--table-header-bg: #f1f5f9;
	--row-hover: #f8fafc;
	--muted: #64748b;
	--badge-bg: #e2e8f0;
	--badge-text: #475569;
	--btn-bg: #ffffff;
	--btn-border: #cbd5e1;
	--card-bg: {{if .Style.CardBg}}{{.Style.CardBg}}{{else}}#ffffff{{end}};
}

{{if eq .Style.DarkMode "auto"}}
@media (prefers-color-scheme: dark) {
	:root {
		--bg: #0f172a;
		--text: #f1f5f9;
		--sidebar-bg: #0b1120;
		--sidebar-border: #1e293b;
		--border: #334155;
		--table-header-bg: #1e293b;
		--row-hover: #1e293b;
		--muted: #94a3b8;
		--badge-bg: #1e293b;
		--badge-text: #cbd5e1;
		--btn-bg: #1e293b;
		--btn-border: #334155;
		--card-bg: #0f172a;
	}
}
{{end}}

html.dark {
	--bg: #0f172a;
	--text: #f1f5f9;
	--sidebar-bg: #0b1120;
	--sidebar-border: #1e293b;
	--border: #334155;
	--table-header-bg: #1e293b;
	--row-hover: #1e293b;
	--muted: #94a3b8;
	--badge-bg: #1e293b;
	--badge-text: #cbd5e1;
	--btn-bg: #1e293b;
	--btn-border: #334155;
	--card-bg: #0f172a;
}

/* AG Grid Theme Customization */
.ag-theme-quartz, .ag-theme-quartz-dark, .ag-theme-quartz-auto-dark {
	--ag-font-family: var(--font);
	--ag-accent-color: var(--accent);
	--ag-border-radius: 8px;
	--ag-header-background-color: var(--table-header-bg);
	--ag-header-foreground-color: var(--text);
	--ag-row-hover-color: var(--row-hover);
	--ag-background-color: var(--bg);
	--ag-foreground-color: var(--text);
	--ag-border-color: var(--border);
	--ag-cell-horizontal-border: solid var(--border);
}

* {
	box-sizing: border-box;
}

body {
	margin: 0;
	padding: 0;
	font-family: var(--font);
	background-color: var(--bg);
	color: var(--text);
	display: flex;
	height: 100vh;
	overflow: hidden;
}

/* Sidebar */
.sidebar {
	width: 260px;
	background-color: var(--sidebar-bg);
	border-right: 1px solid var(--sidebar-border);
	display: flex;
	flex-direction: column;
	flex-shrink: 0;
	transition: transform 0.3s ease;
	z-index: 100;
}

.logo {
	padding: 20px;
	font-weight: 700;
	font-size: 1.15rem;
	border-bottom: 1px solid var(--sidebar-border);
	display: flex;
	align-items: center;
	gap: 10px;
	color: var(--text);
}

.nav-search {
	padding: 12px 16px;
	border-bottom: 1px solid var(--sidebar-border);
}

.nav-search input {
	width: 100%;
	padding: 8px 12px;
	border: 1px solid var(--border);
	border-radius: 6px;
	background-color: var(--bg);
	color: var(--text);
	font-family: var(--font);
	font-size: 0.875rem;
	outline: none;
	transition: border-color 0.2s;
}

.nav-search input:focus {
	border-color: var(--accent);
}

.nav-list {
	list-style: none;
	padding: 8px 0;
	margin: 0;
	overflow-y: auto;
	flex-grow: 1;
}

.nav-item {
	margin: 0;
}

.nav-link {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 10px 18px;
	text-decoration: none;
	color: var(--text);
	border-left: 3px solid transparent;
	font-size: 0.9rem;
	transition: background-color 0.2s, border-color 0.2s;
}

.nav-link:hover {
	background-color: var(--row-hover);
}

.nav-link.active {
	border-left-color: var(--accent);
	background-color: var(--row-hover);
	font-weight: 600;
	color: var(--accent);
}

.badge {
	font-size: 0.65rem;
	font-weight: 600;
	padding: 2px 6px;
	border-radius: 6px;
	background-color: var(--badge-bg);
	color: var(--badge-text);
	text-transform: uppercase;
	letter-spacing: 0.5px;
}

.row-count {
	font-size: 0.75rem;
	color: var(--muted);
	background-color: var(--border);
	padding: 2px 6px;
	border-radius: 10px;
}

/* Main Content */
main {
	flex-grow: 1;
	overflow-y: auto;
	padding: 32px 40px;
	scroll-behavior: smooth;
}

section {
	margin-bottom: 50px;
	padding-bottom: 20px;
	border-bottom: 1px solid var(--border);
}

section:last-child {
	border-bottom: none;
}

.section-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	flex-wrap: wrap;
	gap: 12px;
	margin-bottom: 14px;
}

h2 {
	margin: 0;
	font-size: 1.5rem;
	font-weight: 600;
	display: flex;
	align-items: center;
	gap: 10px;
}

.table-meta {
	font-size: 0.85rem;
	color: var(--muted);
	display: flex;
	gap: 12px;
	align-items: center;
}

/* AG Grid Wrapper & Toolbar */
.ag-grid-wrapper {
	margin-top: 12px;
	border-radius: 8px;
	box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.grid-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	flex-wrap: wrap;
	gap: 10px;
	padding: 10px 14px;
	background-color: var(--table-header-bg);
	border: 1px solid var(--border);
	border-bottom: none;
	border-top-left-radius: 8px;
	border-top-right-radius: 8px;
}

.toolbar-left, .toolbar-right {
	display: flex;
	align-items: center;
	gap: 8px;
}

.grid-search-input {
	padding: 6px 12px;
	border: 1px solid var(--btn-border);
	border-radius: 6px;
	background-color: var(--bg);
	color: var(--text);
	font-family: var(--font);
	font-size: 0.825rem;
	min-width: 220px;
	outline: none;
}

.grid-search-input:focus {
	border-color: var(--accent);
}

.tool-btn {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	padding: 6px 12px;
	font-size: 0.825rem;
	font-weight: 500;
	font-family: var(--font);
	color: var(--text);
	background-color: var(--btn-bg);
	border: 1px solid var(--btn-border);
	border-radius: 6px;
	cursor: pointer;
	transition: all 0.15s ease-in-out;
}

.tool-btn:hover {
	border-color: var(--accent);
	color: var(--accent);
}

.grid-stats {
	font-size: 0.8rem;
	color: var(--muted);
}

.ag-grid-box {
	border: 1px solid var(--border);
	border-bottom-left-radius: 8px;
	border-bottom-right-radius: 8px;
	overflow: hidden;
}

/* Fallback Static Table */
.fallback-table-wrapper {
	border: 1px solid var(--border);
	border-radius: 8px;
	overflow-x: auto;
	background-color: var(--bg);
	margin-top: 12px;
}

table.fallback-table {
	width: 100%;
	border-collapse: collapse;
	text-align: left;
	font-size: 0.875rem;
}

table.fallback-table th, table.fallback-table td {
	padding: 10px 14px;
	border-bottom: 1px solid var(--border);
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	max-width: 300px;
}

table.fallback-table th {
	background-color: var(--table-header-bg);
	font-weight: 600;
	color: var(--muted);
	position: sticky;
	top: 0;
	z-index: 10;
}

/* Sticky first column for wide fallback tables */
table.fallback-table.wide-table th:first-child,
table.fallback-table.wide-table td:first-child {
	position: sticky;
	left: 0;
	background-color: var(--bg);
	z-index: 11;
	border-right: 1px solid var(--border);
}

table.fallback-table.wide-table th:first-child {
	background-color: var(--table-header-bg);
	z-index: 12;
}

table.fallback-table tr:hover td {
	background-color: var(--row-hover);
}

.fallback-pagination {
	display: flex;
	justify-content: center;
	align-items: center;
	margin-top: 14px;
	gap: 12px;
}

.fallback-page-btn {
	background-color: var(--btn-bg);
	border: 1px solid var(--btn-border);
	color: var(--text);
	padding: 5px 14px;
	border-radius: 6px;
	cursor: pointer;
	font-size: 0.85rem;
}

.fallback-page-btn:disabled {
	opacity: 0.5;
	cursor: not-allowed;
}

/* Mobile Header */
.mobile-header {
	display: none;
	padding: 14px 20px;
	background-color: var(--sidebar-bg);
	border-bottom: 1px solid var(--border);
	align-items: center;
	justify-content: space-between;
}

.hamburger {
	background: none;
	border: none;
	color: var(--text);
	font-size: 1.5rem;
	cursor: pointer;
}

@media (max-width: 768px) {
	body {
		flex-direction: column;
	}
	.mobile-header {
		display: flex;
	}
	.sidebar {
		position: absolute;
		top: 55px;
		left: 0;
		bottom: 0;
		transform: translateX(-100%);
	}
	.sidebar.open {
		transform: translateX(0);
	}
	main {
		padding: 20px 16px;
	}
}

/* Custom Injected CSS from _head */
{{if .Head.CustomCSS}}
{{safeCSS .Head.CustomCSS}}
{{end}}
</style>

<!-- Progressive enhancement / noscript styles -->
<noscript>
	<style>
		.ag-grid-wrapper { display: none !important; }
		.fallback-table-wrapper { display: block !important; }
	</style>
</noscript>
</head>
<body>

<div class="mobile-header">
	<div style="font-weight: bold">{{if .Style.LogoText}}{{.Style.LogoText}} {{end}}{{if .Head.Title}}{{.Head.Title}}{{else}}{{.Style.Title}}{{end}}</div>
	<button class="hamburger" onclick="document.querySelector('.sidebar').classList.toggle('open')">☰</button>
</div>

<nav class="sidebar">
	<div class="logo">
		{{if .Style.LogoText}}<span>{{.Style.LogoText}}</span>{{end}}
		<span>{{if .Head.Title}}{{.Head.Title}}{{else}}{{.Style.Title}}{{end}}</span>
	</div>
	<div class="nav-search">
		<input type="text" id="searchInput" placeholder="Filter tables..." onkeyup="filterNav()">
	</div>
	<ul class="nav-list" id="navList">
		{{range .Tables}}
		{{if not .Hidden}}
		<li class="nav-item">
			<a href="#table-{{slugify .Name}}" class="nav-link" data-target="table-{{slugify .Name}}">
				<div>
					{{.Label}}
					<span class="badge">{{.Type}}</span>
				</div>
				<span class="row-count">{{.RowCount}}</span>
			</a>
		</li>
		{{end}}
		{{end}}
	</ul>
</nav>

<main id="main">
	{{range .Contents}}
	<section id="table-{{slugify .Info.Name}}">
		<div class="section-header">
			<h2>
				{{.Info.Label}}
				<span class="badge">{{.Info.Type}}</span>
			</h2>
			<div class="table-meta">
				<span>{{.Info.RowCount}} rows</span>
				<span>•</span>
				<span>{{.Info.ColCount}} columns</span>
			</div>
		</div>

		<!-- Embedded JSON Data for AG Grid -->
		<script type="application/json" id="data-{{slugify .Info.Name}}">
			{{toGridJSON .}}
		</script>

		<!-- Interactive AG Grid Container (Activated when JS is available) -->
		<div class="ag-grid-wrapper" id="ag-grid-wrapper-{{slugify .Info.Name}}" style="display: none;">
			<div class="grid-toolbar">
				<div class="toolbar-left">
					<input type="text" 
						class="grid-search-input" 
						placeholder="Search in table..." 
						oninput="onQuickFilterChanged('{{slugify .Info.Name}}', this.value)"
					>
					<span class="grid-stats" id="grid-stats-{{slugify .Info.Name}}">{{.Info.RowCount}} rows</span>
				</div>
				<div class="toolbar-right">
					<button class="tool-btn" onclick="autoFitColumns('{{slugify .Info.Name}}')">
						<span>↔</span> Auto-Fit
					</button>
					<button class="tool-btn" onclick="resetGridFilters('{{slugify .Info.Name}}')">
						<span>↺</span> Reset
					</button>
					<button class="tool-btn" onclick="exportGridCsv('{{slugify .Info.Name}}', '{{.Info.Name}}')">
						<span>⤓</span> Export CSV
					</button>
				</div>
			</div>
			<div class="ag-grid-box">
				<div id="grid-{{slugify .Info.Name}}" 
					class="{{if eq $.Style.DarkMode "dark"}}ag-theme-quartz-dark{{else if eq $.Style.DarkMode "light"}}ag-theme-quartz{{else}}ag-theme-quartz-auto-dark{{end}}" 
					style="height: {{if gt .Info.RowCount 15}}540px{{else}}{{add 120 (mul .Info.RowCount 42)}}px{{end}}; width: 100%;">
				</div>
			</div>
		</div>

		<!-- Fallback Static HTML Table (Graceful degradation for non-JS environments) -->
		<div class="fallback-table-wrapper" id="fallback-{{slugify .Info.Name}}">
			<table class="fallback-table {{if gt .Info.ColCount 10}}wide-table{{end}}" id="data-table-{{slugify .Info.Name}}">
				<thead>
					<tr>
						{{range .Columns}}
						<th title="{{.}}">{{.}}</th>
						{{end}}
					</tr>
				</thead>
				<tbody>
					{{$psize := .PageSize}}
					{{range $rowIndex, $row := .Rows}}
					{{$pageNum := pageForRow $rowIndex $psize}}
					<tr data-page="{{$pageNum}}" style="{{if ne $pageNum 1}}display: none;{{end}}">
						{{range $row}}
						<td title="{{.}}">{{.}}</td>
						{{end}}
					</tr>
					{{end}}
				</tbody>
			</table>

			{{if gt .TotalPages 1}}
			<div class="fallback-pagination">
				<button class="fallback-page-btn prev-btn" onclick="changeFallbackPage('{{slugify .Info.Name}}', -1)" disabled>← Previous</button>
				<span class="fallback-page-info" id="fallback-page-info-{{slugify .Info.Name}}">Page 1 of {{.TotalPages}}</span>
				<button class="fallback-page-btn next-btn" onclick="changeFallbackPage('{{slugify .Info.Name}}', 1)">Next →</button>
			</div>
			{{end}}
		</div>

	</section>
	{{end}}
</main>

<!-- Embedded Minified AG Grid Community Library -->
<script>
{{embedAGGridJS}}
</script>

<!-- Application & AG Grid Initialization Script -->
<script>
const gridInstances = {};

function changeFallbackPage(tableId, delta) {
	const table = document.getElementById('data-table-' + tableId);
	if (!table) return;
	let currentPage = parseInt(table.getAttribute('data-current-page') || '1');
	const totalPages = parseInt(table.getAttribute('data-total-pages') || '1');
	
	let newPage = currentPage + delta;
	if (newPage < 1 || newPage > totalPages) return;
	
	table.setAttribute('data-current-page', newPage);
	const rows = table.querySelectorAll('tbody tr');
	rows.forEach(row => {
		row.style.display = (row.getAttribute('data-page') == newPage) ? '' : 'none';
	});
	
	const section = table.closest('section');
	if (section) {
		const prevBtn = section.querySelector('.prev-btn');
		const nextBtn = section.querySelector('.next-btn');
		const info = section.querySelector('.fallback-page-info');
		if (prevBtn) prevBtn.disabled = newPage === 1;
		if (nextBtn) nextBtn.disabled = newPage === totalPages;
		if (info) info.textContent = 'Page ' + newPage + ' of ' + totalPages;
	}
}

function filterNav() {
	const query = document.getElementById('searchInput').value.toLowerCase();
	const items = document.querySelectorAll('.nav-item');
	items.forEach(item => {
		const text = item.textContent.toLowerCase();
		item.style.display = text.includes(query) ? '' : 'none';
	});
}

function onQuickFilterChanged(tableId, value) {
	const grid = gridInstances[tableId];
	if (grid) {
		grid.setGridOption('quickFilterText', value);
	}
}

function exportGridCsv(tableId, fileName) {
	const grid = gridInstances[tableId];
	if (grid) {
		grid.exportDataAsCsv({ fileName: (fileName || 'export') + '.csv' });
	}
}

function resetGridFilters(tableId) {
	const grid = gridInstances[tableId];
	if (grid) {
		grid.setFilterModel(null);
		grid.setGridOption('quickFilterText', '');
		const searchInput = document.querySelector('#ag-grid-wrapper-' + tableId + ' .grid-search-input');
		if (searchInput) searchInput.value = '';
	}
}

function autoFitColumns(tableId) {
	const grid = gridInstances[tableId];
	if (grid) {
		grid.autoSizeAllColumns();
	}
}

document.addEventListener('DOMContentLoaded', function() {
	if (typeof agGrid === 'undefined' || !agGrid.createGrid) {
		console.warn('AG Grid library not available; running in static HTML table fallback mode.');
		return;
	}

	const sections = document.querySelectorAll('main section');
	sections.forEach(section => {
		const slugId = section.getAttribute('id').replace('table-', '');
		const scriptTag = document.getElementById('data-' + slugId);
		const gridDiv = document.getElementById('grid-' + slugId);
		const agWrapper = document.getElementById('ag-grid-wrapper-' + slugId);
		const fallbackWrapper = document.getElementById('fallback-' + slugId);

		if (!scriptTag || !gridDiv || !agWrapper) return;

		try {
			const data = JSON.parse(scriptTag.textContent);
			const colCount = data.colCount || (data.columns ? data.columns.length : 0);
			const rowCount = data.rowCount || (data.rows ? data.rows.length : 0);

			const colDefs = data.columns.map((colName, index) => {
				const isWide = colCount > 10;
				const def = {
					field: colName,
					headerName: colName,
					filter: true,
					floatingFilter: true,
					sortable: true,
					resizable: true,
					tooltipField: colName,
					minWidth: 120,
				};

				if (isWide && index === 0) {
					def.pinned = 'left';
					def.lockPinned = true;
					def.cellStyle = { fontWeight: '600' };
				}

				if (colCount <= 6) {
					def.flex = 1;
				}

				return def;
			});

			const pageSize = data.pageSize || 25;
			const gridOptions = {
				rowData: data.rows,
				columnDefs: colDefs,
				pagination: true,
				paginationPageSize: pageSize,
				paginationPageSizeSelector: [10, 25, 50, 100, 250, 500],
				enableCellTextSelection: true,
				ensureDomOrder: true,
				rowHoverHighlight: true,
				animateRows: true,
				multiSortKey: 'ctrl',
				accentedSort: true,
				enableBrowserTooltips: true,
				defaultColDef: {
					sortable: true,
					filter: true,
					floatingFilter: true,
					resizable: true,
				},
				onFilterChanged: (event) => {
					const stats = document.getElementById('grid-stats-' + slugId);
					if (stats && event.api) {
						const displayed = event.api.getDisplayedRowCount();
						stats.textContent = displayed + ' of ' + rowCount + ' rows';
					}
				},
				onGridReady: (params) => {
					if (colCount > 6) {
						params.api.autoSizeAllColumns();
					} else {
						params.api.sizeColumnsToFit();
					}
				}
			};

			const gridApi = agGrid.createGrid(gridDiv, gridOptions);
			gridInstances[slugId] = gridApi;

			agWrapper.style.display = 'block';
			if (fallbackWrapper) {
				fallbackWrapper.style.display = 'none';
			}

		} catch (err) {
			console.error('Failed to initialize AG Grid for table ' + slugId + ', falling back to HTML table:', err);
			if (fallbackWrapper) fallbackWrapper.style.display = 'block';
			if (agWrapper) agWrapper.style.display = 'none';
		}
	});

	const main = document.getElementById('main');
	const navLinks = document.querySelectorAll('.nav-link');

	if (main && navLinks.length > 0) {
		main.addEventListener('scroll', () => {
			let current = '';
			sections.forEach(section => {
				const sectionTop = section.offsetTop;
				if (main.scrollTop >= sectionTop - 120) {
					current = section.getAttribute('id');
				}
			});

			navLinks.forEach(link => {
				link.classList.remove('active');
				if (link.getAttribute('data-target') === current) {
					link.classList.add('active');
				}
			});
		});

		navLinks.forEach(link => {
			link.addEventListener('click', function(e) {
				e.preventDefault();
				const targetId = this.getAttribute('data-target');
				const targetSection = document.getElementById(targetId);
				document.querySelector('.sidebar').classList.remove('open');
				if (targetSection) {
					main.scrollTo({
						top: targetSection.offsetTop - 20,
						behavior: 'smooth'
					});
				}
			});
		});

		const firstId = sections[0]?.getAttribute('id');
		const firstLink = document.querySelector('.nav-link[data-target="' + firstId + '"]');
		if (firstLink) firstLink.classList.add('active');
	}
});
</script>
</body>
</html>
`
