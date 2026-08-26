# sqldoc Examples & Usage

### 1. Create the Demo Database
```sh
sqlite3 demo.db < create_demo.sql
```

### 2. Build the Tools
```sh
# Build all tools into bin/
go build -o ../bin/sqldoc ../cmd/sqldoc
go build -o ../bin/sqldoc-serve ../cmd/sqldoc-serve
go build -o ../bin/sqldoc-viewer ../cmd/sqldoc-viewer
```

### 3. Usage Options

#### Option A: CLI HTML Export (`sqldoc`)
Render to a standalone HTML file with embedded AG Grid:
```sh
../bin/sqldoc render demo.db -o output.html
open output.html
```

Inspect database metadata and prioritized table order:
```sh
../bin/sqldoc info demo.db
```

#### Option B: Local Web Server & Browser Launcher (`sqldoc-serve`)
Start a local HTTP service that renders the database in real-time and opens your browser:
```sh
../bin/sqldoc-serve demo.db
```
*Optional: specify port `-p 8080`*

#### Option C: Native WebKit Desktop Viewer (`sqldoc-viewer`)
Open the database in a self-contained native desktop window (using macOS WKWebView / system WebKit) with zero browser or network dependencies:
```sh
../bin/sqldoc-viewer demo.db
```
