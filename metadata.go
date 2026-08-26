package sqldoc

import (
	"database/sql"
	"strconv"
	"strings"
)

// MetaTag represents an arbitrary <meta> element in <head>
type MetaTag struct {
	Name      string
	Property  string
	HttpEquiv string
	Content   string
}

// LinkTag represents an arbitrary <link> element in <head>
type LinkTag struct {
	Rel         string
	Href        string
	Type        string
	As          string
	CrossOrigin string
}

// StyleConfig represents styling properties (which can be defined in _head or _style)
type StyleConfig struct {
	Title       string
	AccentColor string
	BgColor     string
	TextColor   string
	FontFamily  string
	PageSize    int    // rows per page, default 100
	LogoText    string
	DarkMode    string // auto | light | dark
	CardBg      string
	SidebarBg   string
}

// HeadConfig represents all HTML <head> metadata
type HeadConfig struct {
	// Document title
	Title string

	// Base tag
	BaseHref   string
	BaseTarget string

	// Standard Meta
	Charset     string
	Viewport    string
	Description string
	Keywords    string
	Author      string
	Robots      string
	Generator   string
	ThemeColor  string
	ColorScheme string

	// OpenGraph Meta
	OGTitle       string
	OGDescription string
	OGImage       string
	OGUrl         string
	OGType        string
	OGSiteName    string

	// Twitter Meta
	TwitterCard        string
	TwitterTitle       string
	TwitterDescription string
	TwitterImage       string
	TwitterSite        string
	TwitterCreator     string

	// Arbitrary Custom Metas & Links
	CustomMetas []MetaTag
	CustomLinks []LinkTag

	// Resource Links
	Favicon        string
	AppleTouchIcon string
	Canonical      string
	Stylesheets    []string
	Preconnects    []string

	// Scripts & Custom Injections
	HeadScripts []string
	CustomCSS   string
	RawHead     string

	// Embedded Style Config subset
	Style StyleConfig
}

// DefaultStyle returns default styling options
func DefaultStyle() StyleConfig {
	return StyleConfig{
		Title:       "SQLite Document",
		AccentColor: "#2563eb",
		BgColor:     "#ffffff",
		TextColor:   "#1e293b",
		FontFamily:  "system-ui, -apple-system, sans-serif",
		PageSize:    100,
		LogoText:    "📊",
		DarkMode:    "auto",
	}
}

// DefaultHead returns default <head> metadata
func DefaultHead() HeadConfig {
	style := DefaultStyle()
	return HeadConfig{
		Title:       style.Title,
		Charset:     "UTF-8",
		Viewport:    "width=device-width, initial-scale=1.0",
		Generator:   "sqldoc (SQLite Document Engine)",
		ColorScheme: "light dark",
		Style:       style,
	}
}

// LoadHead loads all head metadata and style properties from _head (and/or _style)
func LoadHead(db *sql.DB) HeadConfig {
	head := DefaultHead()

	// 1. Check and read from _style table (if present) to initialize style subset
	loadStyleTable(db, &head)

	// 2. Check and read from _head table (primary metadata table)
	loadHeadTable(db, &head)

	// Synchronize Title between Head and Style if one was set
	if head.Title != "" && head.Style.Title == DefaultStyle().Title {
		head.Style.Title = head.Title
	} else if head.Style.Title != "" && head.Title == DefaultHead().Title {
		head.Title = head.Style.Title
	}

	return head
}

// LoadStyle is a backwards-compatible helper that returns the Style subset
func LoadStyle(db *sql.DB) StyleConfig {
	return LoadHead(db).Style
}

func loadStyleTable(db *sql.DB, head *HeadConfig) {
	var exists string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='_style'").Scan(&exists)
	if err != nil || exists == "" {
		return
	}

	rows, err := db.Query("SELECT key, value FROM _style")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		applyHeadKeyValue(head, key, value)
	}
}

func loadHeadTable(db *sql.DB, head *HeadConfig) {
	var exists string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='_head'").Scan(&exists)
	if err != nil || exists == "" {
		return
	}

	// Check table columns to support flexible schemas
	pragmaRows, err := db.Query(`PRAGMA table_info("_head")`)
	if err != nil {
		return
	}
	cols := make(map[string]bool)
	for pragmaRows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}
		if err := pragmaRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
			cols[strings.ToLower(name)] = true
		}
	}
	pragmaRows.Close()

	// If standard key-value or name-content schema:
	if (cols["key"] && cols["value"]) || (cols["name"] && cols["content"]) {
		keyCol := "key"
		valCol := "value"
		if !cols["key"] {
			keyCol = "name"
			valCol = "content"
		}
		query := "SELECT \"" + keyCol + "\", \"" + valCol + "\" FROM _head"
		rows, err := db.Query(query)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err == nil {
				applyHeadKeyValue(head, key, value)
			}
		}
		return
	}

	// If generic multi-column schema (tag, name, property, content, href, rel, etc.)
	rows, err := db.Query("SELECT * FROM _head")
	if err != nil {
		return
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return
	}

	colIdx := make(map[string]int)
	for i, c := range colNames {
		colIdx[strings.ToLower(c)] = i
	}

	for rows.Next() {
		vals := make([]interface{}, len(colNames))
		scanArgs := make([]interface{}, len(colNames))
		for i := range vals {
			scanArgs[i] = &vals[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}

		getString := func(col string) string {
			idx, ok := colIdx[col]
			if !ok || vals[idx] == nil {
				return ""
			}
			if b, ok := vals[idx].([]byte); ok {
				return string(b)
			}
			return vals[idx].(string)
		}

		tag := strings.ToLower(getString("tag"))
		name := getString("name")
		prop := getString("property")
		content := getString("content")
		href := getString("href")
		rel := getString("rel")
		key := getString("key")
		value := getString("value")

		if key != "" && value != "" {
			applyHeadKeyValue(head, key, value)
			continue
		}

		if tag == "meta" || name != "" || prop != "" {
			head.CustomMetas = append(head.CustomMetas, MetaTag{
				Name:      name,
				Property:  prop,
				HttpEquiv: getString("http_equiv"),
				Content:   content,
			})
		} else if tag == "link" || rel != "" || href != "" {
			head.CustomLinks = append(head.CustomLinks, LinkTag{
				Rel:         rel,
				Href:        href,
				Type:        getString("type"),
				As:          getString("as"),
				CrossOrigin: getString("crossorigin"),
			})
		} else if tag == "title" {
			head.Title = content
			head.Style.Title = content
		} else if tag == "style" {
			head.CustomCSS += "\n" + content
		} else if tag == "script" {
			if href != "" {
				head.HeadScripts = append(head.HeadScripts, href)
			} else if content != "" {
				head.HeadScripts = append(head.HeadScripts, content)
			}
		}
	}
}

// applyHeadKeyValue maps key-value pairs into HeadConfig and StyleConfig
func applyHeadKeyValue(head *HeadConfig, key, value string) {
	k := strings.ToLower(strings.TrimSpace(key))
	v := strings.TrimSpace(value)
	if v == "" {
		return
	}

	switch k {
	// Title
	case "title":
		head.Title = v
		head.Style.Title = v

	// Base
	case "base", "base_href", "base-href":
		head.BaseHref = v
	case "base_target", "base-target":
		head.BaseTarget = v

	// Standard Meta
	case "charset":
		head.Charset = v
	case "viewport":
		head.Viewport = v
	case "description":
		head.Description = v
	case "keywords":
		head.Keywords = v
	case "author":
		head.Author = v
	case "robots":
		head.Robots = v
	case "generator":
		head.Generator = v
	case "theme_color", "theme-color":
		head.ThemeColor = v
	case "color_scheme", "color-scheme":
		head.ColorScheme = v

	// OpenGraph
	case "og:title":
		head.OGTitle = v
	case "og:description":
		head.OGDescription = v
	case "og:image":
		head.OGImage = v
	case "og:url":
		head.OGUrl = v
	case "og:type":
		head.OGType = v
	case "og:site_name", "og:sitename":
		head.OGSiteName = v

	// Twitter
	case "twitter:card":
		head.TwitterCard = v
	case "twitter:title":
		head.TwitterTitle = v
	case "twitter:description":
		head.TwitterDescription = v
	case "twitter:image":
		head.TwitterImage = v
	case "twitter:site":
		head.TwitterSite = v
	case "twitter:creator":
		head.TwitterCreator = v

	// Links & Icons
	case "favicon", "icon", "shortcut_icon":
		head.Favicon = v
	case "apple_touch_icon", "apple-touch-icon":
		head.AppleTouchIcon = v
	case "canonical":
		head.Canonical = v
	case "stylesheet", "css_url", "font_url":
		head.Stylesheets = append(head.Stylesheets, v)
	case "preconnect":
		head.Preconnects = append(head.Preconnects, v)

	// Custom Code & Scripts
	case "custom_css", "style", "css":
		head.CustomCSS += "\n" + v
	case "head_script", "script":
		head.HeadScripts = append(head.HeadScripts, v)
	case "raw_head", "raw_html", "head_html":
		head.RawHead += "\n" + v

	// Style subset properties
	case "accent_color", "accent":
		head.Style.AccentColor = v
	case "bg_color", "bg":
		head.Style.BgColor = v
	case "text_color", "text":
		head.Style.TextColor = v
	case "font_family", "font":
		head.Style.FontFamily = v
	case "page_size", "pagesize":
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			head.Style.PageSize = ps
		}
	case "logo_text", "logo":
		head.Style.LogoText = v
	case "dark_mode", "darkmode":
		head.Style.DarkMode = v
	case "card_bg":
		head.Style.CardBg = v
	case "sidebar_bg":
		head.Style.SidebarBg = v

	// Generic OpenGraph or Twitter or Meta key
	default:
		if strings.HasPrefix(k, "og:") {
			head.CustomMetas = append(head.CustomMetas, MetaTag{Property: key, Content: v})
		} else if strings.HasPrefix(k, "twitter:") {
			head.CustomMetas = append(head.CustomMetas, MetaTag{Name: key, Content: v})
		} else if strings.HasPrefix(k, "meta:") {
			metaName := strings.TrimPrefix(key, "meta:")
			head.CustomMetas = append(head.CustomMetas, MetaTag{Name: metaName, Content: v})
		} else {
			// Store as custom meta tag
			head.CustomMetas = append(head.CustomMetas, MetaTag{Name: key, Content: v})
		}
	}
}
