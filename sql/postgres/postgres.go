package postgres

import (
	"ariga.io/atlas/sql/schema"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Types for PostgreSQL.
var (
// TypeInt is defined in another file in this package
// Uncomment if needed in this file and remove from the other location
// TypeInt = &schema.IntegerType{T: "int"}
)

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// parseTableConstraints parses the table constraints.
func (i *inspect) parseTableConstraints(ctx context.Context, s *schema.Schema, tables []*schema.Table) error {
	rows, err := i.conn.QueryContext(ctx, tableConstraintsQuery, s.Name)
	if err != nil {
		return fmt.Errorf("postgres: querying table constraints: %w", err)
	}
	defer rows.Close()
	// pgIndex represents a PostgreSQL index.
	type pgIndex struct {
		name   string // Index name
		table  string // Table name
		unique bool   // Unique index
	}

	// pgConstraint represents a PostgreSQL constraint.
	type pgConstraint struct {
		name string // Constraint name
		typ  string // Constraint type
	}

	var (
		statistics  = make(map[string][]*pgIndex)
		constraints = make(map[string][]*pgConstraint)
		excludes    = make(map[string][]*pgExcludeConstraint)
	)
	for rows.Next() {
		var tbl, name, typ string
		if err := rows.Scan(&tbl, &name, &typ); err != nil {
			return fmt.Errorf("postgres: scanning index_columns_usage row: %w", err)
		}
		switch typ {
		case "i", "I":
			statistics[tbl] = append(statistics[tbl], &pgIndex{name: name})
		case "p", "u", "f":
			constraints[tbl] = append(constraints[tbl], &pgConstraint{name: name, typ: typ})
		case "x":
			excludes[tbl] = append(excludes[tbl], &pgExcludeConstraint{name: name})
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	// Process the migration specs in a separate function
	// The previous code with t and ts variables was incorrect

	if err := i.setStatisticsInfo(ctx, s, tables, statistics); err != nil {
		return err
	}
	if err := i.setConstraintsInfo(ctx, s, tables, constraints); err != nil {
		return err
	}
	return i.processExcludeConstraints(ctx, s, tables, excludes)
}

// pgExcludeConstraint represents a PostgreSQL EXCLUDE constraint.
type pgExcludeConstraint struct {
	name      string   // Constraint name
	table     string   // Table name
	columns   []string // Column names
	ops       []string // Operators for each column
	predicate string   // WHERE predicate
	using     string   // Index method (e.g., "gist", "btree")
	exprs     []string // Expressions
}

// setStatisticsInfo queries and sets statistics information for tables.
func (i *inspect) setStatisticsInfo(ctx context.Context, s *schema.Schema, tables []*schema.Table, statistics map[string][]*pgIndex) error {
	if len(statistics) == 0 {
		return nil
	}

	// Create a map of tables by name for quick lookup
	tM := make(map[string]*schema.Table, len(tables))
	for _, t := range tables {
		tM[t.Name] = t
	}

	// Query to get index information
	rows, err := i.conn.QueryContext(ctx, indexColumnsQuery, s.Name)
	if err != nil {
		return fmt.Errorf("postgres: querying index columns: %w", err)
	}
	defer rows.Close()

	// Process each row
	for rows.Next() {
		var tname, iname string
		var unique bool
		var cols []string

		if err := rows.Scan(&tname, &iname, &unique, &cols); err != nil {
			return fmt.Errorf("postgres: scanning index row: %w", err)
		}

		// Find the table
		t, ok := tM[tname]
		if !ok {
			continue
		}

		// Create and add the index
		idx := &schema.Index{
			Name:    iname,
			Unique:  unique,
			Columns: cols,
		}

		t.Indexes = append(t.Indexes, idx)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// processExcludeConstraints queries and sets the EXCLUDE constraints information.
func (i *inspect) processExcludeConstraints(ctx context.Context, s *schema.Schema, tables []*schema.Table, excludes map[string][]*pgExcludeConstraint) error {
	if len(excludes) == 0 {
		return nil
	}

	// Create a map of tables by name for quick lookup.
	tableMap := make(map[string]*schema.Table, len(tables))
	for _, t := range tables {
		tableMap[t.Name] = t
	}

	// Query to get EXCLUDE constraint details
	rows, err := i.conn.QueryContext(ctx, excludeConstraintsQuery, s.Name)
	if err != nil {
		return fmt.Errorf("postgres: querying exclude constraints: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tname, cname, using, predicate, colstr, opstr, exprstr string
		)
		if err := rows.Scan(&tname, &cname, &using, &predicate, &colstr, &opstr, &exprstr); err != nil {
			return fmt.Errorf("postgres: scanning exclude constraint row: %w", err)
		}

		t, ok := tableMap[tname]
		if !ok {
			continue
		}

		// Parse columns and operators
		var columns, ops, exprs []string

		if colstr != "" {
			columns = strings.Split(colstr, ",")
		}

		if opstr != "" {
			ops = strings.Split(opstr, ",")
		}

		if exprstr != "" {
			exprs = strings.Split(exprstr, ",")
		}

		// Create and add the EXCLUDE constraint to the table
		ex := &schema.ExcludeConstraint{
			Name:      cname,
			Columns:   columns,
			Ops:       ops,
			Using:     using,
			Predicate: predicate,
			Exprs:     exprs,
		}

		// Add constraint to table attributes
		t.AddAttrs(ex)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// setConstraintsInfo queries and sets the constraint information for tables.
func (i *inspect) setConstraintsInfo(ctx context.Context, s *schema.Schema, tables []*schema.Table, constraints map[string][]*pgConstraint) error {
	if len(constraints) == 0 {
		return nil
	}

	// Create a map of tables by name for quick lookup
	tableMap := make(map[string]*schema.Table, len(tables))
	for _, t := range tables {
		tableMap[t.Name] = t
	}

	// Query constraint details
	rows, err := i.querySchema(ctx, constraintInfoQuery, s)
	if err != nil {
		return fmt.Errorf("postgres: querying constraint info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tname, cname, ctype, refTable, colstr, refColstr string
			updateRule, deleteRule                           string
		)
		if err := rows.Scan(&tname, &cname, &ctype, &refTable, &colstr, &refColstr, &updateRule, &deleteRule); err != nil {
			return fmt.Errorf("postgres: scanning constraint info row: %w", err)
		}

		t, ok := tableMap[tname]
		if !ok {
			continue
		}

		// Parse columns
		columns := strings.Split(colstr, ",")

		switch ctype {
		case "p":
			// Primary key
			pk := schema.NewPrimaryKey(t)
			pk.Name = cname
			for _, c := range columns {
				if col := findColumn(t, c); col != nil {
					pk.AddColumns(*col)
				}
			}
			t.PrimaryKey = pk
		case "u":
			// Unique constraint
			uc := &schema.Index{Name: cname, Unique: true}
			for _, c := range columns {
				if col := findColumn(t, c); col != nil {
					uc.AddColumns(*col)
				}
			}
			t.Indexes = append(t.Indexes, uc)
		case "f":
			// Foreign key
			ref, ok := tableMap[refTable]
			if !ok {
				continue
			}
			fk := &schema.ForeignKey{
				Symbol:   cname,
				Table:    t,
				RefTable: ref,
				OnUpdate: schema.ReferenceOption(updateRule),
				OnDelete: schema.ReferenceOption(deleteRule),
			}

			// Add columns to foreign key
			refColumns := strings.Split(refColstr, ",")
			for i, c := range columns {
				col := findColumn(t, c)
				if col == nil {
					continue
				}
				fk.AddColumns(*col)
				if i < len(refColumns) {
					if refCol := findColumn(ref, refColumns[i]); refCol != nil {
						fk.AddRefColumns(*refCol)
					}
				}
			}
			t.ForeignKeys = append(t.ForeignKeys, fk)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// findColumn finds a column by name in the given table.
func findColumn(t *schema.Table, name string) *schema.Column {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}
	return nil
}

// querySchema is a helper method to query a schema with proper error handling
func (i *inspect) querySchema(ctx context.Context, query string, s interface{}) (*sql.Rows, error) {
	var name string
	switch s := s.(type) {
	case *schema.Schema:
		name = s.Name
	case string:
		name = s
	default:
		return nil, fmt.Errorf("postgres: unexpected schema type: %T", s)
	}
	rows, err := i.conn.QueryContext(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("postgres: executing query: %w", err)
	}
	return rows, nil
}

// Query to fetch index columns information
const indexColumnsQuery = `
SELECT
    c.relname AS table_name,
    i.relname AS index_name,
    i.indisunique AS is_unique,
    array_agg(a.attname ORDER BY k.i) AS column_names
FROM pg_index i
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_class ic ON ic.oid = i.indexrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN LATERAL unnest(i.indkey) WITH ORDINALITY AS k(attnum, i) ON TRUE
JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = k.attnum
WHERE n.nspname = $1 AND i.indisprimary = false
GROUP BY c.relname, i.relname, i.indisunique
ORDER BY c.relname, i.relname
`

// Query to fetch table constraints information
const tableConstraintsQuery = `
SELECT
    c.relname AS table_name,
    i.relname AS index_name,
    CASE
        WHEN i.indisprimary THEN 'p'
        WHEN i.indisunique THEN 'u'
        WHEN con.contype = 'f' THEN 'f'
        WHEN con.contype = 'x' THEN 'x'
        ELSE 'i'
    END AS type
FROM pg_index i
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_class ic ON ic.oid = i.indexrelid
LEFT JOIN pg_constraint con ON con.conindid = i.indexrelid
WHERE c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $1)
  AND i.indislive = true
ORDER BY c.relname, i.indisprimary DESC, i.indisunique DESC, con.contype = 'f' DESC, i.relname
`

// Query to get constraint information
const constraintInfoQuery = `
SELECT
    tc.relname AS table_name,
    c.conname AS constraint_name,
    c.contype AS constraint_type,
    ft.relname AS ref_table,
    (
        SELECT string_agg(a.attname, ',')
        FROM (
            SELECT unnest(c.conkey) AS attnum
        ) AS x
        JOIN pg_attribute a ON a.attnum = x.attnum AND a.attrelid = c.conrelid
        WHERE a.attnum > 0
    ) AS columns,
    (
        SELECT string_agg(a.attname, ',')
        FROM (
            SELECT unnest(c.confkey) AS attnum
        ) AS x
        JOIN pg_attribute a ON a.attnum = x.attnum AND a.attrelid = c.confrelid
        WHERE a.attnum > 0 AND c.confrelid > 0
    ) AS ref_columns,
    CASE c.confupdtype
        WHEN 'a' THEN 'no_action'
        WHEN 'r' THEN 'restrict'
        WHEN 'c' THEN 'cascade'
        WHEN 'n' THEN 'set_null'
        WHEN 'd' THEN 'set_default'
        ELSE 'no_action'
    END AS update_rule,
    CASE c.confdeltype
        WHEN 'a' THEN 'no_action'
        WHEN 'r' THEN 'restrict'
        WHEN 'c' THEN 'cascade'
        WHEN 'n' THEN 'set_null'
        WHEN 'd' THEN 'set_default'
        ELSE 'no_action'
    END AS delete_rule
FROM pg_constraint c
JOIN pg_class tc ON tc.oid = c.conrelid
LEFT JOIN pg_class ft ON ft.oid = c.confrelid
WHERE tc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $1)
  AND c.contype IN ('p', 'u', 'f')
ORDER BY tc.relname, c.contype
`

// Query to get EXCLUDE constraint details including columns, operators, and predicates
const excludeConstraintsQuery = `
SELECT
    tc.relname AS table_name,
    c.conname AS constraint_name,
    am.amname AS using_method,
    pg_get_expr(c.conpred, c.conrelid) AS predicate,
    (
        SELECT string_agg(a.attname, ',')
        FROM (
            SELECT unnest(c.conkey) AS attnum
        ) AS x
        JOIN pg_attribute a ON a.attnum = x.attnum AND a.attrelid = c.conrelid
        WHERE a.attnum > 0
    ) AS columns,
    (
        SELECT string_agg(o.oprname, ',')
        FROM (
            SELECT unnest(c.conexclop) AS oprname
        ) AS x
        JOIN pg_operator o ON o.oid = x.oprname
    ) AS operators,
    (
        SELECT string_agg(pg_get_indexdef(c.conindid, a.attnum, false), ',')
        FROM pg_attribute a
        WHERE a.attrelid = c.conindid AND a.attnum > 0
    ) AS expressions
FROM
    pg_constraint c
JOIN
    pg_class tc ON tc.oid = c.conrelid
JOIN
    pg_index i ON i.indexrelid = c.conindid
JOIN
    pg_class ic ON ic.oid = i.indexrelid
JOIN
    pg_am am ON am.oid = ic.relam
WHERE
    c.contype = 'x' AND tc.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $1)
`

// AddExcludeConstraint builds and executes the query for adding an EXCLUDE constraint.
func (d *driver) AddExcludeConstraint(ctx context.Context, t *schema.Table, c *schema.ExcludeConstraint) error {
	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	b.WriteString(fmt.Sprintf("%q", t.Name))
	b.WriteString(" ADD CONSTRAINT ")
	b.WriteString(fmt.Sprintf("%q", c.Name))
	b.WriteString(" EXCLUDE USING ")
	b.WriteString(c.Using)
	b.WriteString(" (")

	for i := 0; i < max(len(c.Columns), len(c.Exprs)); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		// Add column or expression
		if i < len(c.Exprs) && c.Exprs[i] != "" {
			b.WriteString(c.Exprs[i])
		} else if i < len(c.Columns) {
			b.WriteString(fmt.Sprintf("%q", c.Columns[i]))
		}

		// Add operator
		if i < len(c.Ops) {
			b.WriteString(" WITH ")
			b.WriteString(c.Ops[i])
		}
	}

	b.WriteString(")")

	// Add predicate if exists
	if c.Predicate != "" {
		b.WriteString(" ")
		b.WriteString(c.Predicate)
	}

	_, err := d.ExecContext(ctx, b.String())
	return err
}
