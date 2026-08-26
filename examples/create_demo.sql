CREATE TABLE _head (
    key   TEXT PRIMARY KEY,
    value TEXT
);

-- Document Title & Basic Metadata
INSERT INTO _head VALUES ('title', 'Acme Corp Executive Dashboard');
INSERT INTO _head VALUES ('description', 'Real-time performance analytics, team directory, and product catalog for Acme Corp.');
INSERT INTO _head VALUES ('keywords', 'acme, dashboard, analytics, metrics, revenue, products, employees');
INSERT INTO _head VALUES ('author', 'Acme Corp Data Team');
INSERT INTO _head VALUES ('robots', 'index, follow');
INSERT INTO _head VALUES ('theme_color', '#6366f1');

-- Social & OpenGraph Metadata
INSERT INTO _head VALUES ('og:title', 'Acme Corp Executive Dashboard');
INSERT INTO _head VALUES ('og:description', 'Comprehensive operational metrics and data overview.');
INSERT INTO _head VALUES ('og:type', 'website');
INSERT INTO _head VALUES ('twitter:card', 'summary_large_image');
INSERT INTO _head VALUES ('twitter:title', 'Acme Corp Dashboard');

-- Icons & Stylesheet Links
INSERT INTO _head VALUES ('favicon', 'data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🏢</text></svg>');

-- Style Subset Properties
INSERT INTO _head VALUES ('accent_color', '#6366f1');
INSERT INTO _head VALUES ('bg_color', '#ffffff');
INSERT INTO _head VALUES ('text_color', '#1e293b');
INSERT INTO _head VALUES ('font_family', 'system-ui, -apple-system, sans-serif');
INSERT INTO _head VALUES ('page_size', '25');
INSERT INTO _head VALUES ('logo_text', '🏢');
INSERT INTO _head VALUES ('dark_mode', 'auto');
INSERT INTO _head VALUES ('custom_css', '.logo { letter-spacing: -0.5px; }');

CREATE TABLE home_dashboard (
    metric TEXT,
    value TEXT,
    trend TEXT,
    period TEXT
);

INSERT INTO home_dashboard VALUES ('Revenue', '$1.2M', '+15%', 'Q3');
INSERT INTO home_dashboard VALUES ('Active Users', '45,231', '+5%', 'Q3');
INSERT INTO home_dashboard VALUES ('NPS Score', '72', '+2', 'Q3');
INSERT INTO home_dashboard VALUES ('Churn Rate', '1.2%', '-0.1%', 'Q3');
INSERT INTO home_dashboard VALUES ('MRR', '$400k', '+10%', 'Q3');
INSERT INTO home_dashboard VALUES ('ARR', '$4.8M', '+10%', 'Q3');
INSERT INTO home_dashboard VALUES ('Support Tickets', '1,204', '-5%', 'Q3');
INSERT INTO home_dashboard VALUES ('Uptime', '99.99%', '0%', 'Q3');

CREATE TABLE employees (
    id INTEGER PRIMARY KEY,
    first_name TEXT,
    last_name TEXT,
    email TEXT,
    department TEXT,
    title TEXT,
    salary INTEGER,
    hire_date TEXT,
    location TEXT,
    status TEXT
);

WITH RECURSIVE cnt(x) AS (
  SELECT 1
  UNION ALL
  SELECT x+1 FROM cnt
  LIMIT 120
)
INSERT INTO employees (id, first_name, last_name, email, department, title, salary, hire_date, location, status)
SELECT 
  x, 
  'EmpFirst'||x, 
  'EmpLast'||x, 
  'employee'||x||'@acme.corp', 
  CASE (x % 8)
    WHEN 0 THEN 'Engineering' WHEN 1 THEN 'Marketing' WHEN 2 THEN 'Sales' WHEN 3 THEN 'HR' WHEN 4 THEN 'Finance' WHEN 5 THEN 'Operations' WHEN 6 THEN 'Product' WHEN 7 THEN 'Design' END,
  'Title Level '||((x % 5) + 1),
  55000 + (x * 1234 % 140000),
  date('2018-01-15', '+' || (x * 15) || ' days'),
  CASE (x % 8)
    WHEN 0 THEN 'New York' WHEN 1 THEN 'San Francisco' WHEN 2 THEN 'London' WHEN 3 THEN 'Tokyo' WHEN 4 THEN 'Berlin' WHEN 5 THEN 'Austin' WHEN 6 THEN 'Seattle' WHEN 7 THEN 'Chicago' END,
  CASE (x % 5)
    WHEN 0 THEN 'On Leave' WHEN 1 THEN 'Remote' ELSE 'Active' END
FROM cnt;


CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    name TEXT,
    sku TEXT,
    category TEXT,
    price REAL,
    cost REAL,
    margin REAL,
    stock_qty INTEGER,
    reorder_level INTEGER,
    supplier TEXT,
    weight_kg REAL,
    dimensions TEXT,
    color TEXT,
    material TEXT,
    rating REAL
);

WITH RECURSIVE cnt(x) AS (
  SELECT 1 UNION ALL SELECT x+1 FROM cnt LIMIT 30
)
INSERT INTO products (id, name, sku, category, price, cost, margin, stock_qty, reorder_level, supplier, weight_kg, dimensions, color, material, rating)
SELECT
  x,
  'Product '||x,
  'SKU-'||printf('%04d', x),
  CASE (x%4) WHEN 0 THEN 'Tech' WHEN 1 THEN 'Office' WHEN 2 THEN 'Furniture' WHEN 3 THEN 'Apparel' END,
  10.0 + x*2.5,
  5.0 + x*1.0,
  5.0 + x*1.5,
  100 - x*2,
  20,
  'Supplier '||(x%5),
  1.5 + (x*0.1),
  '10x10x'||(x%10 + 5),
  CASE (x%3) WHEN 0 THEN 'Red' WHEN 1 THEN 'Blue' WHEN 2 THEN 'Black' END,
  'Plastic',
  4.0 + (x%10)*0.1
FROM cnt;

CREATE VIEW quarterly_results AS
SELECT department, COUNT(*) as employee_count, SUM(salary) as total_salary, ROUND(AVG(salary), 2) as avg_salary
FROM employees
GROUP BY department;

CREATE TABLE orders (
    id INTEGER PRIMARY KEY,
    customer_name TEXT,
    product_id INTEGER,
    quantity INTEGER,
    total REAL,
    order_date TEXT,
    status TEXT,
    shipping_method TEXT
);

WITH RECURSIVE cnt(x) AS (
  SELECT 1 UNION ALL SELECT x+1 FROM cnt LIMIT 50
)
INSERT INTO orders (id, customer_name, product_id, quantity, total, order_date, status, shipping_method)
SELECT
  x,
  'Customer '||x,
  (x % 30) + 1,
  (x % 5) + 1,
  100.50 * ((x % 5) + 1),
  date('2024-01-01', '+' || (x * 2) || ' days'),
  CASE (x % 4) WHEN 0 THEN 'Pending' WHEN 1 THEN 'Shipped' WHEN 2 THEN 'Delivered' WHEN 3 THEN 'Cancelled' END,
  CASE (x % 3) WHEN 0 THEN 'Standard' WHEN 1 THEN 'Express' WHEN 2 THEN 'Overnight' END
FROM cnt;

CREATE TABLE _nav (
    table_name TEXT PRIMARY KEY,
    label      TEXT,
    position   INTEGER,
    hidden     INTEGER DEFAULT 0
);
INSERT INTO _nav VALUES ('home_dashboard', 'Dashboard', 1, 0);
INSERT INTO _nav VALUES ('employees', 'Team Directory', 2, 0);
INSERT INTO _nav VALUES ('quarterly_results', 'Dept Summary', 3, 0);
