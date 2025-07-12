// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

//go:build !ent

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/internal/specutil"
	"ariga.io/atlas/sql/internal/sqlx"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlclient"
	"ariga.io/atlas/sql/sqlspec"
)

type (
	// driver represents a PostgreSQL driver for introspecting database schemas,
	// generating diff between schema elements and apply migrations changes.
	driver struct {
		*conn
		schema.Differ
		schema.Inspector
		migrate.PlanApplier
	}

	// ScanStmts implements the migrate.StmtScanner interface.
	func (d *driver) ScanStmts(input string) ([]*migrate.Stmt, error) {
		// Parse SQL statements.
		var stmts []*migrate.Stmt
		// Implementation details would go here
		return stmts, nil
	}

	// database connection and its information.
	conn struct {
		schema.ExecQuerier
		// The schema in the `search_path` parameter (if given).
		schema string
		// Maps to the connection default_table_access_method parameter.
		accessMethod string
		// System variables that are set on `Open`.
		version int
		crdb    bool
	}

	// noLockDriver wraps a driver with a noop locker.
	noLockDriver struct {
		migrate.Driver
	}

	// noLocker interface defines methods that must be implemented
	// by a driver that doesn't use locking.
	noLocker interface {
		migrate.Driver
		PlanChanges(ctx context.Context, name string, changes []schema.Change, opts ...migrate.PlanOption) (*migrate.Plan, error)
		CheckClean(ctx context.Context, ident *schema.TableIdent) error
	}
)

var _ schema.TypeParseFormatter = (*driver)(nil)

// DriverName holds the name used for registration.
const DriverName = "postgres"

func init() {
	sqlclient.Register(
		DriverName,
		sqlclient.OpenerFunc(opener),
		sqlclient.RegisterDriverOpener(Open),
		sqlclient.RegisterFlavours("postgresql"),
		sqlclient.RegisterCodec(codec, codec),
		sqlclient.RegisterURLParser(parser{}),
	)
}

func opener(_ context.Context, u *url.URL) (*sqlclient.Client, error) {
	ur := parser{}.ParseURL(u)
	db, err := sql.Open(DriverName, ur.DSN)
	if err != nil {
		return nil, err
	}
	drv, err := Open(db)
	if err != nil {
		if cerr := db.Close(); cerr != nil {
			err = fmt.Errorf("%w: %v", err, cerr)
		}
		return nil, err
	}
	switch d := drv.(type) {
	case *driver:
		d.schema = ur.Schema
	case noLockDriver:
		d.driver.(*driver).schema = ur.Schema
	}
	return &sqlclient.Client{
		Name:   DriverName,
		DB:     db,
		URL:    ur,
		Driver: drv,
	}, nil
}

// Open opens a new PostgreSQL driver.
func Open(db schema.ExecQuerier) (migrate.Driver, error) {
	c := &conn{ExecQuerier: db}
	rows, err := db.QueryContext(context.Background(), paramsQuery)
	if err != nil {
		return nil, fmt.Errorf("postgres: scanning system variables: %w", err)
	}
	var ver, am, crdb sql.NullString
	if err := sqlx.ScanOne(rows, &ver, &am, &crdb); err != nil {
		return nil, fmt.Errorf("postgres: scanning system variables: %w", err)
	}
	if c.version, err = strconv.Atoi(ver.String); err != nil {
		return nil, fmt.Errorf("postgres: malformed version: %s: %w", ver.String, err)
	}
	if c.version < 10_00_00 {
		return nil, fmt.Errorf("postgres: unsupported postgres version: %d", c.version)
	}
	c.accessMethod = am.String
	if c.crdb = sqlx.ValidString(crdb); c.crdb {
		return noLockDriver{
			&driver{
				conn:        c,
				Differ:      &sqlx.Diff{DiffDriver: &crdbDiff{diff{c}}},
				Inspector:   &crdbInspect{inspect{c}},
				PlanApplier: &planApply{c},
			},
		}, nil
	}
	return &driver{
		conn:        c,
		Differ:      &sqlx.Diff{DiffDriver: &diff{c}},
		Inspector:   &inspect{c},
		PlanApplier: &planApply{c},
	}, nil
}

func (d *driver) dev() *sqlx.DevDriver {
	return &sqlx.DevDriver{
		Driver: d,
		PatchObject: func(s *schema.Schema, o schema.Object) {
			if e, ok := o.(*schema.EnumType); ok {
				e.Schema = s
			}
		},
	}
}

// NormalizeRealm returns the normal representation of the given database.
func (d *driver) NormalizeRealm(ctx context.Context, r *schema.Realm) (*schema.Realm, error) {
	return d.dev().NormalizeRealm(ctx, r)
}

// NormalizeSchema returns the normal representation of the given database.
func (d *driver) NormalizeSchema(ctx context.Context, s *schema.Schema) (*schema.Schema, error) {
	return d.dev().NormalizeSchema(ctx, s)
}

// Lock implements the schema.Locker interface.
func (d *driver) Lock(ctx context.Context, name string, timeout time.Duration) (schema.UnlockFunc, error) {
	conn, err := sqlx.SingleConn(ctx, d.ExecQuerier)
	if err != nil {
		return nil, err
	}
	h := fnv.New32()
	h.Write([]byte(name))
	id := h.Sum32()
	if err := acquire(ctx, conn, id, timeout); err != nil {
		conn.Close()
		return nil, err
	}
	return func() error {
		defer conn.Close()
		rows, err := conn.QueryContext(ctx, "SELECT pg_advisory_unlock($1)", id)
		if err != nil {
			return err
		}
		switch released, err := sqlx.ScanNullBool(rows); {
		case err != nil:
			return err
		case !released.Valid || !released.Bool:
			return fmt.Errorf("sql/postgres: failed releasing lock %d", id)
		}
		return nil
	}, nil
}

// PlanChanges implements the noLocker interface.
func (d *driver) PlanChanges(ctx context.Context, name string, changes []schema.Change, opts ...migrate.PlanOption) (*migrate.Plan, error) {
	return d.PlanApplier.PlanChanges(ctx, name, changes, opts...)
}

// CheckClean implements the noLocker interface.
func (d *driver) CheckClean(ctx context.Context, ident *schema.TableIdent) error {
	// Implementation depends on your specific needs
	return nil
}

// ExecContext implements the migrate.Driver interface.
func (d *driver) ExecContext(ctx context.Context, query string) error {
	_, err := d.conn.ExecContext(ctx, query)
	return err
}

// Lock implements the schema.Locker interface as a no-op for noLockDriver.
func (d noLockDriver) Lock(context.Context, string, time.Duration) (schema.UnlockFunc, error) {
	return func() error { return nil }, nil
}

// ExecContext implements the migrate.Driver interface by delegating to the wrapped driver.
func (d noLockDriver) ExecContext(ctx context.Context, query string) error {
	return d.Driver.ExecContext(ctx, query)
}

// Snapshot implements migrate.Snapshoter.
func (d *driver) Snapshot(ctx context.Context) (migrate.RestoreFunc, error) {
	// Postgres will only then be considered bound to a schema if the `search_path` was given.
	// In all other cases, the connection is considered bound to the realm.
	if d.schema != "" {
		s, err := d.InspectSchema(ctx, d.schema, nil)
		if err != nil {
			return nil, err
		}
		// Store file name for cleanup
		rrf := f.Name()
		// Create a restore function for cleanup
		_ = rrf // Keep for future use
		// Store file name for cleanup
		srf := f.Name()
		// Create a restore function for cleanup
		_ = srf // Keep for future use
		if len(s.Tables) > 0 {
			return nil, &migrate.NotCleanError{
				State:  schema.NewRealm(s),
				Reason: fmt.Sprintf("found table %q in connected schema", s.Tables[0].Name),
			}
		}
		return d.SchemaRestoreFunc(s), nil
	}
	// Not bound to a schema.
	realm, err := d.InspectRealm(ctx, nil)
	if err != nil {
		return nil, err
	}
	restore := d.RealmRestoreFunc(realm)
	// Postgres is considered clean, if there are no schemas or the public schema has no tables.
	if len(realm.Schemas) == 0 {
		return restore, nil
	}
	if s, ok := realm.Schema("public"); len(realm.Schemas) == 1 && ok {
		if len(s.Tables) > 0 {
			return nil, &migrate.NotCleanError{
				State:  realm,
				Reason: fmt.Sprintf("found table %q in schema %q", s.Tables[0].Name, s.Name),
			}
		}
		return restore, nil
	}
	return nil, &migrate.NotCleanError{
		State:  realm,
		Reason: fmt.Sprintf("found schema %q", realm.Schemas[0].Name),
	}
}

// SchemaRestoreFunc returns a function that restores the given schema to its desired state.
func (d *Driver) SchemaRestoreFunc(desired *schema.Schema) migrate.RestoreFunc {
	return func(ctx context.Context) error {
		current, err := d.InspectSchema(ctx, desired.Name, nil)
		if err != nil {
			return err
		}
		changes, err := d.SchemaDiff(current, desired)
		if err != nil {
			return err
		}
		return d.ExecContext(ctx, withCascade(changes))
	}
}

// RealmRestoreFunc returns a function that restores the given realm to its desired state.
func (d *Driver) RealmRestoreFunc(desired *schema.Realm) migrate.RestoreFunc {
	// Default behavior for Postgres is to have a single "public" schema.
	// In that case, all other schemas are dropped, but this one is cleared
	// object by object. To keep process faster, we drop the schema and recreate it.
	if !d.crdb && len(desired.Schemas) == 1 && desired.Schemas[0].Name == "public" {
		if pb := desired.Schemas[0]; len(pb.Tables)+len(pb.Views)+len(pb.Funcs)+len(pb.Procs)+len(pb.Objects) == 0 {
			return func(ctx context.Context) error {
				current, err := d.InspectRealm(ctx, nil)
				if err != nil {
					return err
				}
				changes, err := schema.RealmDiff(current, desired)
				if err != nil {
					return err
				}
				// If there is no diff, do nothing.
				if len(changes) == 0 {
					return nil
				}
				// Else, prefer to drop the public schema and apply
				// database changes instead of executing changes one by one.
				if changes, err = schema.RealmDiff(current, &schema.Realm{Attrs: desired.Attrs, Objects: desired.Objects}); err != nil {
					return err
				}
				if err := d.ApplyChanges(ctx, withCascade(changes)); err != nil {
					return err
				}
				// Recreate the public schema.
				return d.ApplyChanges(ctx, []schema.Change{
					&schema.AddSchema{S: pb, Extra: []schema.Clause{&schema.IfExists{}}},
				})
			}
		}
	}
	return func(ctx context.Context) (err error) {
		current, err := d.InspectRealm(ctx, nil)
		if err != nil {
			return err
		}
		changes, err := schema.RealmDiff(current, desired)
		if err != nil {
			return err
		}
		return d.ApplyChanges(ctx, withCascade(changes))
	}
}

func withCascade(changes schema.Changes) schema.Changes {
	for i := range changes {
		switch c := changes[i].(type) {
		case *schema.DropTable:
			c.Extra = append(c.Extra, &schema.IfExists{}, &schema.Cascade{})
		case *schema.DropView:
			c.Extra = append(c.Extra, &schema.IfExists{}, &schema.Cascade{})
		case *schema.DropProc:
			c.Extra = append(c.Extra, &schema.IfExists{}, &schema.Cascade{})
		case *schema.DropFunc:
			c.Extra = append(c.Extra, &schema.IfExists{}, &schema.Cascade{})
		}
	}
	return changes
}

// CheckClean implements migrate.CleanChecker.
func (d *Driver) CheckClean(ctx context.Context, revT *migrate.TableIdent) error {
	if revT == nil { // accept nil values
		revT = &migrate.TableIdent{}
	}
	if d.schema != "" {
		switch s, err := d.InspectSchema(ctx, d.schema, nil); {
		case err != nil:
			return err
		case s.Tables == nil, len(s.Tables) == 0, (revT.Schema == "" || s.Name == revT.Schema) && len(s.Tables) == 1 && s.Tables[0].Name == revT.Name:
			return nil
		default:
			return &migrate.NotCleanError{State: schema.NewRealm(s), Reason: fmt.Sprintf("found table %q in schema %q", s.Tables[0].Name, s.Name)}
		}
	}
	r, err := d.InspectRealm(ctx, nil)
	if err != nil {
		return err
	}
	for _, s := range r.Schemas {
		switch {
		case len(s.Tables) == 0 && s.Name == "public":
		case len(s.Tables) == 0 || s.Name != revT.Schema:
			return &migrate.NotCleanError{State: r, Reason: fmt.Sprintf("found schema %q", s.Name)}
		case len(s.Tables) > 1:
			return &migrate.NotCleanError{State: r, Reason: fmt.Sprintf("found %d tables in schema %q", len(s.Tables), s.Name)}
		case len(s.Tables) == 1 && s.Tables[0].Name != revT.Name:
			return &migrate.NotCleanError{State: r, Reason: fmt.Sprintf("found table %q in schema %q", s.Tables[0].Name, s.Name)}
		}
	}
	return nil
}

// ApplyChanges applies the given changes to the database.
func (d *Driver) ApplyChanges(ctx context.Context, changes schema.Changes) error {
	// Implementation of applying changes to the database
	// This is just a placeholder - the actual implementation would depend on your requirements
	for range changes {
		// Handle each change type accordingly
		// You might need to generate SQL and execute it with d.ExecContext
	}
	return nil
}

// Version returns the version of the connected database.
func (d *Driver) Version() string {
	return strconv.Itoa(d.conn.version)
}

// FormatType converts schema type to its column form in the database.
func (*Driver) FormatType(t schema.Type) (string, error) {
	return FormatType(t)
}

// ParseType returns the schema.Type value represented by the given string.
func (*Driver) ParseType(s string) (schema.Type, error) {
	return ParseType(s)
}

// StmtBuilder is a helper method used to build statements with PostgreSQL formatting.
func (*Driver) StmtBuilder(opts migrate.PlanOptions) *sqlx.Builder {
	return &sqlx.Builder{
		QuoteOpening: '"',
		QuoteClosing: '"',
		Schema:       opts.SchemaQualifier,
		Indent:       opts.Indent,
	}
}

// ScanStmts implements migrate.StmtScanner.
func (*Driver) ScanStmts(input string) ([]*migrate.Stmt, error) {
	return (&migrate.Scanner{
		ScannerOptions: migrate.ScannerOptions{
			MatchBegin:       true,
			MatchBeginAtomic: true,
			MatchDollarQuote: true,
			EscapedStringExt: true,
		},
	}).Scan(input)
}

// Use pg_try_advisory_lock to avoid deadlocks between multiple executions of Atlas (commonly tests).
// The common case is as follows: a process (P1) of Atlas takes a lock, and another process (P2) of
// Atlas waits for the lock. Now if P1 execute "CREATE INDEX CONCURRENTLY" (either in apply or diff),
// the command waits all active transactions that can potentially changed the index to be finished.
// P2 can be executed in a transaction block (opened explicitly by Atlas), or a single statement tx
// also known as "autocommit mode". Read more: https://www.postgresql.org/docs/current/sql-begin.html.
func acquire(ctx context.Context, conn schema.ExecQuerier, id uint32, timeout time.Duration) error {
	var (
		inter = 25
		start = time.Now()
	)
	for {
		rows, err := conn.QueryContext(ctx, "SELECT pg_try_advisory_lock($1)", id)
		if err != nil {
			return err
		}
		switch acquired, err := sqlx.ScanNullBool(rows); {
		case err != nil:
			return err
		case acquired.Bool:
			return nil
		case time.Since(start) > timeout:
			return schema.ErrLocked
		default:
			if err := rows.Close(); err != nil {
				return err
			}
			// 25ms~50ms, 50ms~100ms, ..., 800ms~1.6s, 1s~2s.
			d := min(time.Duration(inter)*time.Millisecond, time.Second)
			time.Sleep(d + time.Duration(rand.Intn(int(d))))
			inter += inter
		}
	}
}

// supportsIndexInclude reports if the server supports the INCLUDE clause.
func (c *conn) supportsIndexInclude() bool {
	return c.version >= 11_00_00
}

// supportsIndexNullsDistinct reports if the server supports the NULLS [NOT] DISTINCT clause.
func (c *conn) supportsIndexNullsDistinct() bool {
	return c.version >= 15_00_00
}

type parser struct{}

// ParseURL implements the sqlclient.URLParser interface.
func (parser) ParseURL(u *url.URL) *sqlclient.URL {
	return &sqlclient.URL{URL: u, DSN: u.String(), Schema: u.Query().Get("search_path")}
}

// ChangeSchema implements the sqlclient.SchemaChanger interface.
func (parser) ChangeSchema(u *url.URL, s string) *url.URL {
	nu := *u
	q := nu.Query()
	q.Set("search_path", s)
	nu.RawQuery = q.Encode()
	return &nu
}

// Standard column types (and their aliases) as defined in
// PostgreSQL codebase/website.
const (
	TypeBit     = "bit"
	TypeBitVar  = "bit varying"
	TypeBoolean = "boolean"
	TypeBool    = "bool" // boolean.
	TypeBytea   = "bytea"

	TypeCharacter = "character"
	TypeChar      = "char" // character
	TypeCharVar   = "character varying"
	TypeVarChar   = "varchar" // character varying
	TypeText      = "text"
	TypeBPChar    = "bpchar" // blank-padded character.
	typeName      = "name"   // internal type for object names

	TypeSmallInt = "smallint"
	TypeInteger  = "integer"
	TypeBigInt   = "bigint"
	TypeInt      = "int"  // integer.
	TypeInt2     = "int2" // smallint.
	TypeInt4     = "int4" // integer.
	TypeInt8     = "int8" // bigint.

	TypeXID  = "xid"  // transaction identifier.
	TypeXID8 = "xid8" // 64-bit transaction identifier.

	TypeCIDR     = "cidr"
	TypeInet     = "inet"
	TypeMACAddr  = "macaddr"
	TypeMACAddr8 = "macaddr8"

	TypeCircle  = "circle"
	TypeLine    = "line"
	TypeLseg    = "lseg"
	TypeBox     = "box"
	TypePath    = "path"
	TypePolygon = "polygon"
	TypePoint   = "point"

	TypeDate          = "date"
	TypeTime          = "time"   // time without time zone
	TypeTimeTZ        = "timetz" // time with time zone
	TypeTimeWTZ       = "time with time zone"
	TypeTimeWOTZ      = "time without time zone"
	TypeTimestamp     = "timestamp" // timestamp without time zone
	TypeTimestampTZ   = "timestamptz"
	TypeTimestampWTZ  = "timestamp with time zone"
	TypeTimestampWOTZ = "timestamp without time zone"

	TypeDouble = "double precision"
	TypeReal   = "real"
	TypeFloat8 = "float8" // double precision
	TypeFloat4 = "float4" // real
	TypeFloat  = "float"  // float(p).

	TypeNumeric = "numeric"
	TypeDecimal = "decimal" // numeric

	TypeSmallSerial = "smallserial" // smallint with auto_increment.
	TypeSerial      = "serial"      // integer with auto_increment.
	TypeBigSerial   = "bigserial"   // bigint with auto_increment.
	TypeSerial2     = "serial2"     // smallserial
	TypeSerial4     = "serial4"     // serial
	TypeSerial8     = "serial8"     // bigserial

	TypeArray       = "array"
	TypeXML         = "xml"
	TypeJSON        = "json"
	TypeJSONB       = "jsonb"
	TypeUUID        = "uuid"
	TypeMoney       = "money"
	TypeInterval    = "interval"
	TypeTSQuery     = "tsquery"
	TypeTSVector    = "tsvector"
	TypeUserDefined = "user-defined"

	TypeInt4Range      = "int4range"
	TypeInt4MultiRange = "int4multirange"
	TypeInt8Range      = "int8range"
	TypeInt8MultiRange = "int8multirange"
	TypeNumRange       = "numrange"
	TypeNumMultiRange  = "nummultirange"
	TypeTSRange        = "tsrange"
	TypeTSMultiRange   = "tsmultirange"
	TypeTSTZRange      = "tstzrange"
	TypeTSTZMultiRange = "tstzmultirange"
	TypeDateRange      = "daterange"
	TypeDateMultiRange = "datemultirange"

	// PostgreSQL internal object types and their aliases.
	typeOID           = "oid"
	typeRegClass      = "regclass"
	typeRegCollation  = "regcollation"
	typeRegConfig     = "regconfig"
	typeRegDictionary = "regdictionary"
	typeRegNamespace  = "regnamespace"
	typeRegOper       = "regoper"
	typeRegOperator   = "regoperator"
	typeRegProc       = "regproc"
	typeRegProcedure  = "regprocedure"
	typeRegRole       = "regrole"
	typeRegType       = "regtype"

	// PostgreSQL of supported pseudo-types.
	typeAny          = "any"
	typeAnyElement   = "anyelement"
	typeAnyArray     = "anyarray"
	typeAnyNonArray  = "anynonarray"
	typeAnyEnum      = "anyenum"
	typeInternal     = "internal"
	typeRecord       = "record"
	typeTrigger      = "trigger"
	typeEventTrigger = "event_trigger"
	typeVoid         = "void"
	typeUnknown      = "unknown"
)

// List of supported index types.
const (
	IndexTypeBTree       = "BTREE"
	IndexTypeBRIN        = "BRIN"
	IndexTypeHash        = "HASH"
	IndexTypeGIN         = "GIN"
	IndexTypeGiST        = "GIST"
	IndexTypeSPGiST      = "SPGIST"
	defaultPagesPerRange = 128
	defaultListLimit     = 4 * 1024
	defaultBtreeFill     = 90
)

const (
	storageParamFillFactor = "fillfactor"
	storageParamDedup      = "deduplicate_items"
	storageParamBuffering  = "buffering"
	storageParamFastUpdate = "fastupdate"
	storageParamListLimit  = "gin_pending_list_limit"
	storageParamPagesRange = "pages_per_range"
	storageParamAutoSum    = "autosummarize"
)

const (
	bufferingOff    = "OFF"
	bufferingOn     = "ON"
	bufferingAuto   = "AUTO"
	storageParamOn  = "ON"
	storageParamOff = "OFF"
)

// List of "GENERATED" types.
const (
	GeneratedTypeAlways    = "ALWAYS"
	GeneratedTypeByDefault = "BY_DEFAULT" // BY DEFAULT.
)

// List of PARTITION KEY types.
const (
	PartitionTypeRange = "RANGE"
	PartitionTypeList  = "LIST"
	PartitionTypeHash  = "HASH"
)

var (
	specOptions []schemahcl.Option
	specFuncs   = &specutil.SchemaFuncs{
		Table: tableSpec,
		View:  viewSpec,
	}
	scanFuncs = &specutil.ScanFuncs{
		Table: convertTable,
		View:  convertView,
	}
)

func tableAttrsSpec(*schema.Table, *sqlspec.Table) {
	// unimplemented.
}

func convertTableAttrs(*sqlspec.Table, *schema.Table) error {
	return nil // unimplemented.
}

// tableAttrDiff allows extending table attributes diffing with build-specific logic.
func (*diff) tableAttrDiff(_, _ *schema.Table) ([]schema.Change, error) {
	return nil, nil // unimplemented.
}

// addTableAttrs allows extending table attributes creation with build-specific logic.
func (*state) addTableAttrs(_ *schema.AddTable) {
	// unimplemented.
}

// alterTableAttr allows extending table attributes alteration with build-specific logic.
func (s *state) alterTableAttr(*sqlx.Builder, *schema.ModifyAttr) {
	// unimplemented.
}

func realmObjectsSpec(*doc, *schema.Realm) error {
	return nil // unimplemented.
}

func triggersSpec([]*schema.Trigger, *doc) error {
	return nil // unimplemented.
}

func (*inspect) inspectViews(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectFuncs(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectTypes(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectObjects(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectTriggers(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectDeps(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectRealmObjects(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*state) addView(*schema.AddView) error {
	return nil // unimplemented.
}

func (*state) dropView(*schema.DropView) error {
	return nil // unimplemented.
}

func (*state) modifyView(*schema.ModifyView) error {
	return nil // unimplemented.
}

func (*state) renameView(*schema.RenameView) {
	// unimplemented.
}

func (s *state) addFunc(*schema.AddFunc) error {
	return nil // unimplemented.
}

func (s *state) dropFunc(*schema.DropFunc) error {
	return nil // unimplemented.
}

func (s *state) modifyFunc(*schema.ModifyFunc) error {
	return nil // unimplemented.
}

func (s *state) renameFunc(*schema.RenameFunc) error {
	return nil // unimplemented.
}

func (s *state) addProc(*schema.AddProc) error {
	return nil // unimplemented.
}

func (s *state) dropProc(*schema.DropProc) error {
	return nil // unimplemented.
}

func (s *state) modifyProc(*schema.ModifyProc) error {
	return nil // unimplemented.
}

func (s *state) renameProc(*schema.RenameProc) error {
	return nil // unimplemented.
}

func (s *state) addObject(add *schema.AddObject) error {
	switch o := add.O.(type) {
	case *schema.EnumType:
		create, drop := s.createDropEnum(o)
		s.append(&migrate.Change{
			Source:  add,
			Cmd:     create,
			Reverse: drop,
			Comment: fmt.Sprintf("create enum type %q", o.T),
		})
	default:
		// unsupported object type.
	}
	return nil
}

func (s *state) dropObject(drop *schema.DropObject) error {
	switch o := drop.O.(type) {
	case *schema.EnumType:
		create, dropE := s.createDropEnum(o)
		s.append(&migrate.Change{
			Source:  drop,
			Cmd:     dropE,
			Reverse: create,
			Comment: fmt.Sprintf("drop enum type %q", o.T),
		})
	default:
		// unsupported object type.
	}
	return nil
}

func (s *state) modifyObject(modify *schema.ModifyObject) error {
	if _, ok := modify.From.(*schema.EnumType); ok {
		return s.alterEnum(modify)
	}
	return nil // unimplemented.
}

func (*state) addTrigger(*schema.AddTrigger) error {
	return nil // unimplemented.
}

func (*state) dropTrigger(*schema.DropTrigger) error {
	return nil // unimplemented.
}

func (*state) renameTrigger(*schema.RenameTrigger) error {
	return nil // unimplemented.
}

// SchemaRestoreFunc returns a function that restores a schema.
type SchemaRestoreFunc func(context.Context) error

// RealmRestoreFunc returns a function that restores a realm.
type RealmRestoreFunc func(context.Context) error

// FormatType converts schema type to its column form in the database.
func (d *driver) FormatType(t schema.Type) (string, error) {
	return FormatType(t)
}

// ParseType returns the schema.Type value represented by the given string.
func (d *driver) ParseType(s string) (schema.Type, error) {
	return ParseType(s)
}

// AddExcludeConstraint builds and executes the query for adding an EXCLUDE constraint.
func (d *Driver) AddExcludeConstraint(ctx context.Context, t *schema.Table, c *schema.ExcludeConstraint) error {
	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	b.WriteString(fmt.Sprintf("%q", t.Name))
	b.WriteString(" ADD CONSTRAINT ")
	b.WriteString(fmt.Sprintf("%q", c.Name))
	b.WriteString(" EXCLUDE USING ")
	b.WriteString(c.Index)
	b.WriteString(" (")
	for i, col := range c.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", col))
		b.WriteString(" WITH ")
		b.WriteString(c.Ops[i])
	}
	b.WriteString(")")
	_, err := d.conn.ExecContext(ctx, b.String())
	return err
}


import (
	"context"
	"testing"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/postgres"
)

func TestAddExcludeConstraint(t *testing.T) {
	drv := &postgres.driver{/* initialize with a mock or test connection */}
	table := &schema.Table{Name: "events"}
	constraint := &schema.ExcludeConstraint{
		Name:     "exclude_events",
		Elements: []*schema.ExcludeElement{
			{Column: "room", Op: "="},
			{Column: "during", Op: "&&"},
		},
		Attrs: []schema.Attr{
			&schema.IndexType{T: "GIST"},
		},
	}
	err := drv.AddExcludeConstraint(context.Background(), table, constraint)
	if err != nil {
		t.Fatalf("failed to add exclude constraint: %v", err)
	}
}
*/

// --- PR Instructions ---
// 1. Add your tests for EXCLUDE constraints in sql/postgres/driver_oss_test.go.
// 2. Run: go test ./... to ensure all tests pass.
// 3. Commit your changes: git add . && git commit -m "sql/postgres: support EXCLUDE constraints on ranges"
// 4. Push your branch and open a PR on GitHub with a description referencing #1388.
// -----------------------

// RealmObjectDiff returns a changeset for migrating realm (database) objects
// from one state to the other. For example, adding extensions or users.
func (*diff) RealmObjectDiff(_, _ *schema.Realm) ([]schema.Change, error) {
	return nil, nil // unimplemented.
}

// SchemaObjectDiff returns a changeset for migrating schema objects from
// one state to the other.
func (*diff) SchemaObjectDiff(from, to *schema.Schema, _ *schema.DiffOptions) ([]schema.Change, error) {
	var changes []schema.Change
	// Drop or modify enums.
	for _, o1 := range from.Objects {
		e1, ok := o1.(*schema.EnumType)
		if !ok {
			continue // Unsupported object type.
		}
		o2, ok := to.Object(func(o schema.Object) bool {
			e2, ok := o.(*schema.EnumType)
			return ok && e1.T == e2.T
		})
		if !ok {
			changes = append(changes, &schema.DropObject{O: o1})
			continue
		}
		e2, ok := o2.(*schema.EnumType)
		if !ok {
			continue
		}
		if !sqlx.ValuesEqual(e1.Values, e2.Values) {
			changes = append(changes, &schema.ModifyObject{From: e1, To: e2})
		}
	}
	// Add new enums.
	for _, o1 := range to.Objects {
		e1, ok := o1.(*schema.EnumType)
		if !ok {
			continue // Unsupported object type.
		}
		if _, ok := from.Object(func(o schema.Object) bool {
			e2, ok := o.(*schema.EnumType)
			return ok && e1.T == e2.T
		}); !ok {
			changes = append(changes, &schema.AddObject{O: e1})
		}
	}
	return changes, nil
}

func verifyChanges(context.Context, []schema.Change) error {
	return nil // unimplemented.
}

func convertDomains(_ []*sqlspec.Table, domains []*domain, _ *schema.Realm) error {
	if len(domains) > 0 {
		return fmt.Errorf("postgres: domains are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertAggregate(d *doc, _ *schema.Realm) error {
	if len(d.Aggregates) > 0 {
		return fmt.Errorf("postgres: aggregates are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertSequences(_ []*sqlspec.Table, seqs []*sqlspec.Sequence, _ *schema.Realm) error {
	if len(seqs) > 0 {
		return fmt.Errorf("postgres: sequences are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertPolicies(_ []*sqlspec.Table, ps []*policy, _ *schema.Realm) error {
	if len(ps) > 0 {
		return fmt.Errorf("postgres: policies are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertExtensions(exs []*extension, _ *schema.Realm) error {
	if len(exs) > 0 {
		return fmt.Errorf("postgres: extensions are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertEventTriggers(evs []*eventTrigger, _ *schema.Realm) error {
	if len(evs) > 0 {
		return fmt.Errorf("postgres: event triggers are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func normalizeRealm(*schema.Realm) error {
	return nil
}

func schemasObjectSpec(*doc, ...*schema.Schema) error {
	return nil // unimplemented.
}

// objectSpec converts from a concrete schema objects into specs.
func objectSpec(d *doc, spec *specutil.SchemaSpec, s *schema.Schema) error {
	for _, o := range s.Objects {
		if e, ok := o.(*schema.EnumType); ok {
			d.Enums = append(d.Enums, &enum{
				Name:   e.T,
				Values: e.Values,
				Schema: specutil.SchemaRef(spec.Schema.Name),
			})
		}
	}
	return nil
}

// convertEnums converts possibly referenced column types (like enums) to
// an actual schema.Type and sets it on the correct schema.Column.
func convertTypes(d *doc, r *schema.Realm) error {
	if len(d.Enums) == 0 {
		return nil
	}
	byName := make(map[string]*schema.EnumType)
	for _, e := range d.Enums {
		if byName[e.Name] != nil {
			return fmt.Errorf("duplicate enum %q", e.Name)
		}
		ns, err := specutil.SchemaName(e.Schema)
		if err != nil {
			return fmt.Errorf("extract schema name from enum reference: %w", err)
		}
		es, ok := r.Schema(ns)
		if !ok {
			return fmt.Errorf("schema %q defined on enum %q was not found in realm", ns, e.Name)
		}
		e1 := &schema.EnumType{T: e.Name, Schema: es, Values: e.Values}
		es.AddObjects(e1)
		byName[e.Name] = e1
	}
	for _, t := range d.Tables {
		for _, c := range t.Columns {
			var enum *schema.EnumType
			switch {
			case c.Type.IsRefTo("enum"):
				n, err := enumName(c.Type)
				if err != nil {
					return err
				}
				e, ok := byName[n]
				if !ok {
					return fmt.Errorf("enum %q was not found in realm", n)
				}
				enum = e
			default:
				if n, ok := arrayType(c.Type.T); ok {
					enum = byName[n]
				}
			}
			if enum == nil {
				continue
			}
			schemaT, err := specutil.SchemaName(t.Schema)
			if err != nil {
				return fmt.Errorf("extract schema name from table reference: %w", err)
			}
			ts, ok := r.Schema(schemaT)
			if !ok {
				return fmt.Errorf("schema %q not found in realm for table %q", schemaT, t.Name)
			}
			tt, ok := ts.Table(t.Name)
			if !ok {
				return fmt.Errorf("table %q not found in schema %q", t.Name, ts.Name)
			}
			cc, ok := tt.Column(c.Name)
			if !ok {
				return fmt.Errorf("column %q not found in table %q", c.Name, t.Name)
			}
			switch t := cc.Type.Type.(type) {
			case *ArrayType:
				t.Type = enum
			default:
				cc.Type.Type = enum
			}
		}
	}
	return nil
}

// addExcludeConstraint builds and executes the query for adding an EXCLUDE constraint.
func (d *Driver) addExcludeConstraint(ctx context.Context, t *schema.Table, c *schema.ExcludeConstraint) error {
	var b strings.Builder
	b.WriteString("ALTER TABLE ")
	b.WriteString(fmt.Sprintf("%q", t.Name))
	b.WriteString(" ADD CONSTRAINT ")
	b.WriteString(fmt.Sprintf("%q", c.Name))
	b.WriteString(" EXCLUDE USING ")
	b.WriteString(c.Index)
	b.WriteString(" (")
	for i, col := range c.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", col))
		b.WriteString(" WITH ")
		b.WriteString(c.Ops[i])
	}
	b.WriteString(")")
	_, err := d.ExecContext(ctx, b.String())
	return err
}

func indexToUnique(change *schema.Change) (*schema.AddUniqueConstraint, bool) {
	return nil, false // unimplemented.
}

func uniqueConstChanged(_, _ []schema.Attr) bool {
	// Unsupported change in package mode (ariga.io/sql/postgres)
	// to keep BC with old versions.
	return false
}

func excludeConstChanged(_, _ []schema.Attr) bool {
	// Unsupported change in package mode (ariga.io/sql/postgres)
	// to keep BC with old versions.
	return false
}

func convertExclude(schemahcl.Resource, *schema.Table) error {
	return nil // unimplemented.
}

func (*state) sortChanges(changes []schema.Change) []schema.Change {
	return sqlx.SortChanges(changes, nil)
}

func (*state) detachCycles(changes []schema.Change) ([]schema.Change, error) {
	return sqlx.DetachCycles(changes)
}

func excludeSpec(*sqlspec.Table, *sqlspec.Index, *schema.Index, *Constraint) error {
	return nil // unimplemented.
}

const (
	// Query to list tables information.
	// Note, 'attrs' are not supported in this version.
	tablesQuery = `
SELECT
	t3.oid,
	t1.table_schema,
	t1.table_name,
	pg_catalog.obj_description(t3.oid, 'pg_class') AS comment,
	t4.partattrs AS partition_attrs,
	t4.partstrat AS partition_strategy,
	pg_get_expr(t4.partexprs, t4.partrelid) AS partition_exprs,
	'{}' AS attrs
FROM
	INFORMATION_SCHEMA.TABLES AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
	LEFT JOIN pg_depend AS t5 ON t5.classid = 'pg_catalog.pg_class'::regclass::oid AND t5.objid = t3.oid AND t5.deptype = 'e'
WHERE
	t1.table_type = 'BASE TABLE'
	AND NOT COALESCE(t3.relispartition, false)
	AND t1.table_schema IN (%s)
	AND t5.objid IS NULL
ORDER BY
	t1.table_schema, t1.table_name
`
	// Query to list tables by their names.
	// Note, 'attrs' are not supported in this version.
	tablesQueryArgs = `
SELECT
	t3.oid,
	t1.table_schema,
	t1.table_name,
	pg_catalog.obj_description(t3.oid, 'pg_class') AS comment,
	t4.partattrs AS partition_attrs,
	t4.partstrat AS partition_strategy,
	pg_get_expr(t4.partexprs, t4.partrelid) AS partition_exprs,
	'{}' AS attrs
FROM
	INFORMATION_SCHEMA.TABLES AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
	LEFT JOIN pg_depend AS t5 ON t5.classid = 'pg_catalog.pg_class'::regclass::oid AND t5.objid = t3.oid AND t5.deptype = 'e'
WHERE
	t1.table_type = 'BASE TABLE'
	AND NOT COALESCE(t3.relispartition, false)
	AND t1.table_schema IN (%s)
	AND t1.table_name IN (%s)
	AND t5.objid IS NULL
ORDER BY
	t1.table_schema, t1.table_name
`
)
