// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"reflect"
	"strconv"
	"strings"
)

// Schema represents a database schema.
type Schema struct {
	Name    string
	Tables  []*Table
	Views   []*View
	Funcs   []*Func
	Procs   []*Proc
	Objects []Object
	Attrs   []Attr
}

// Table represents a database table.
type Table struct {
	Name        string
	Schema      *Schema
	Columns     []*Column
	Indexes     []*Index
	PrimaryKey  *Index
	ForeignKeys []*ForeignKey
	Attrs       []Attr     // Attrs, constraints and options.
	Triggers    []*Trigger // Triggers on the table.
	Deps        []Object   // Objects this table depends on.
	Refs        []Object   // Objects that depends on this table.
}

// Column represents a table column.
type Column struct {
	Name  string
	Type  Type
	Attrs []Attr
}

// Rows is the interface that wraps the basic methods for database rows.
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

// Result is the result of a query execution.
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

type (
	Realm struct {
		Schemas []*Schema
		Attrs   []Attr
		Objects []Object
	}

	Object interface {
		obj()
	}

	View struct {
		Name     string
		Def      string
		Schema   *Schema
		Columns  []*Column
		Attrs    []Attr
		Indexes  []*Index
		Triggers []*Trigger
		Deps     []Object
		Refs     []Object
	}

	NamedDefault struct {
		Expr
		Name  string
		Attrs []Attr
	}

	ColumnType struct {
		Type Type
		Raw  string
		Null bool
	}

	Index struct {
		Name   string
		Unique bool
		Table  *Table
		View   *View
		Attrs  []Attr
		Parts  []*IndexPart
	}

	IndexPart struct {
		SeqNo int
		Desc  bool
		X     Expr
		C     *Column
		Attrs []Attr
	}

	ForeignKey struct {
		Symbol     string
		Table      *Table
		Columns    []*Column
		RefTable   *Table
		RefColumns []*Column
		OnUpdate   ReferenceOption
		OnDelete   ReferenceOption
		Attrs      []Attr
	}

	Trigger struct {
		Name       string
		Table      *Table
		View       *View
		ActionTime TriggerTime
		Events     []TriggerEvent
		For        TriggerFor
		Body       string
		Attrs      []Attr
		Deps       []Object
		Refs       []Object
	}

	TriggerTime string
	TriggerFor  string

	TriggerEvent struct {
		Name    string
		Columns []*Column
	}

	Func struct {
		Name   string
		Schema *Schema
		Args   []*FuncArg
		Ret    Type
		Body   string
		Lang   string
		Attrs  []Attr
		Deps   []Object
		Refs   []Object
	}

	Proc struct {
		Name   string
		Schema *Schema
		Args   []*FuncArg
		Body   string
		Lang   string
		Attrs  []Attr
		Deps   []Object
		Refs   []Object
	}

	FuncArg struct {
		Name    string
		Type    Type
		Default Expr
		Mode    FuncArgMode
		Attrs   []Attr
	}

	FuncArgMode string
)

// List of supported function argument modes.
const (
	FuncArgModeIn       FuncArgMode = "IN"
	FuncArgModeOut      FuncArgMode = "OUT"
	FuncArgModeInOut    FuncArgMode = "INOUT"
	FuncArgModeVariadic FuncArgMode = "VARIADIC"
)

// List of supported trigger action times.
const (
	TriggerTimeBefore  TriggerTime = "BEFORE"
	TriggerTimeAfter   TriggerTime = "AFTER"
	TriggerTimeInstead TriggerTime = "INSTEAD OF"
)

// List of supported trigger FOR EACH spec.
const (
	TriggerForRow  TriggerFor = "ROW"
	TriggerForStmt TriggerFor = "STATEMENT"
)

// List of supported trigger events.
var (
	TriggerEventInsert   = TriggerEvent{Name: "INSERT"}
	TriggerEventUpdate   = TriggerEvent{Name: "UPDATE"}
	TriggerEventDelete   = TriggerEvent{Name: "DELETE"}
	TriggerEventTruncate = TriggerEvent{Name: "TRUNCATE"}
)

// TriggerEventUpdateOf returns an UPDATE OF trigger event.
func TriggerEventUpdateOf(columns ...*Column) TriggerEvent {
	return TriggerEvent{Name: "UPDATE OF", Columns: columns}
}

// Schema returns the first schema that matched the given name.
func (r *Realm) Schema(name string) (*Schema, bool) {
	for _, s := range r.Schemas {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

// Object returns the first object that matched the given predicate.
func (r *Realm) Object(f func(Object) bool) (Object, bool) {
	for _, o := range r.Objects {
		if f(o) {
			return o, true
		}
	}
	return nil, false
}

// PosSetter wraps the two methods for getting and setting positions for schema objects.
type PosSetter interface {
	Pos() *Pos
	SetPos(*Pos)
}

// Pos of the schema, if exists.
func (s *Schema) Pos() *Pos {
	for _, a := range s.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Pos of the enum, if exists.
func (e *EnumType) Pos() *Pos {
	for _, a := range e.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Table returns the first table that matched the given name.
func (s *Schema) Table(name string) (*Table, bool) {
	for _, t := range s.Tables {
		if t.Name == name {
			return t, true
		}
	}
	return nil, false
}

// View returns the first view that matched the given name.
func (s *Schema) View(name string) (*View, bool) {
	for _, v := range s.Views {
		if v.Name == name && !v.Materialized() {
			return v, true
		}
	}
	return nil, false
}

// Materialized returns the first materialized view that matched the given name.
func (s *Schema) Materialized(name string) (*View, bool) {
	for _, v := range s.Views {
		if v.Name == name && v.Materialized() {
			return v, true
		}
	}
	return nil, false
}

// Func returns the first function that matched the given name.
func (s *Schema) Func(name string) (*Func, bool) {
	for _, f := range s.Funcs {
		if f.Name == name {
			return f, true
		}
	}
	return nil, false
}

// Proc returns the first procedure that matched the given name.
func (s *Schema) Proc(name string) (*Proc, bool) {
	for _, p := range s.Procs {
		if p.Name == name {
			return p, true
		}
	}
	return nil, false
}

// Object returns the first object that matched the given predicate.
func (s *Schema) Object(f func(Object) bool) (Object, bool) {
	for _, o := range s.Objects {
		if f(o) {
			return o, true
		}
	}
	return nil, false
}

// Column returns the first column that matched the given name.
func (t *Table) Column(name string) (*Column, bool) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// Pos of the table, if exists.
func (t *Table) Pos() *Pos {
	for _, a := range t.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Index returns the first index that matched the given name.
func (t *Table) Index(name string) (*Index, bool) {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i, true
		}
	}
	return nil, false
}

// ForeignKey returns the first foreign-key that matched the given symbol (constraint name).
func (t *Table) ForeignKey(symbol string) (*ForeignKey, bool) {
	for _, f := range t.ForeignKeys {
		if f.Symbol == symbol {
			return f, true
		}
	}
	return nil, false
}

// Trigger returns the first trigger that matches the given name.
func (t *Table) Trigger(name string) (*Trigger, bool) {
	for _, r := range t.Triggers {
		if r.Name == name {
			return r, true
		}
	}
	return nil, false
}

// Checks of the table.
func (t *Table) Checks() (ck []*Check) {
	for _, a := range t.Attrs {
		if c, ok := a.(*Check); ok {
			ck = append(ck, c)
		}
	}
	return ck
}

// Pos of the view, if exists.
func (v *View) Pos() *Pos {
	for _, a := range v.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Materialized reports if the view is materialized.
func (v *View) Materialized() bool {
	for _, a := range v.Attrs {
		if _, ok := a.(*Materialized); ok {
			return true
		}
	}
	return false
}

// SetMaterialized reports if the view is materialized.
func (v *View) SetMaterialized(b bool) *View {
	if b {
		ReplaceOrAppend(&v.Attrs, &Materialized{})
	} else {
		v.Attrs = RemoveAttr[*Materialized](v.Attrs)
	}
	return v
}

// Column returns the first column that matched the given name.
func (v *View) Column(name string) (*Column, bool) {
	for _, c := range v.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// Index returns the first index that matched the given name.
func (v *View) Index(name string) (*Index, bool) {
	for _, i := range v.Indexes {
		if i.Name == name {
			return i, true
		}
	}
	return nil, false
}

// Trigger returns the first trigger that matches the given name.
func (v *View) Trigger(name string) (*Trigger, bool) {
	for _, r := range v.Triggers {
		if r.Name == name {
			return r, true
		}
	}
	return nil, false
}

// AsTable returns a table that represents the view.
func (v *View) AsTable() *Table {
	return NewTable(v.Name).
		SetSchema(v.Schema).
		AddColumns(v.Columns...)
}

// SetPos sets the position of the function.
func (f *Func) SetPos(p *Pos) {
	ReplaceOrAppend(&f.Attrs, p)
}

// Pos of the schema, if exists.
func (f *Func) Pos() *Pos {
	for _, a := range f.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// SetPos sets the position of the schema.
func (s *Schema) SetPos(p *Pos) {
	ReplaceOrAppend(&s.Attrs, p)
}

// SetPos sets the position of the enum type.
func (e *EnumType) SetPos(p *Pos) {
	ReplaceOrAppend(&e.Attrs, p)
}

// SetPos sets the position of the table.
func (t *Table) SetPos(p *Pos) {
	ReplaceOrAppend(&t.Attrs, p)
}

// SetPos sets the position of the view.
func (v *View) SetPos(p *Pos) {
	ReplaceOrAppend(&v.Attrs, p)
}

// SetPos sets the position of the column.
func (c *Column) SetPos(p *Pos) {
	ReplaceOrAppend(&c.Attrs, p)
}

// SetPos sets the position of the check.
func (c *Check) SetPos(p *Pos) {
	ReplaceOrAppend(&c.Attrs, p)
}

// SetPos sets the position of the index.
func (i *Index) SetPos(p *Pos) {
	ReplaceOrAppend(&i.Attrs, p)
}

// SetPos sets the position of the index part.
func (p *IndexPart) SetPos(p1 *Pos) {
	ReplaceOrAppend(&p.Attrs, p1)
}

// ExcludeConstraint is defined in exclude.go

// SetPos sets the position of the foreign key.
func (f *ForeignKey) SetPos(p *Pos) {
	ReplaceOrAppend(&f.Attrs, p)
}

// SetPos sets the position of the procedure.
func (p *Proc) SetPos(p1 *Pos) {
	ReplaceOrAppend(&p.Attrs, p1)
}

// Pos of the schema, if exists.
func (p *Proc) Pos() *Pos {
	for _, a := range p.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Pos of the column, if exists.
func (c *Column) Pos() *Pos {
	for _, a := range c.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Pos of the index, if exists.
func (i *Index) Pos() *Pos {
	for _, a := range i.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Pos of the index part, if exists.
func (p *IndexPart) Pos() *Pos {
	for _, a := range p.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Pos of the check, if exists.
func (c *Check) Pos() *Pos {
	for _, a := range c.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Pos of the foreign-key, if exists.
func (f *ForeignKey) Pos() *Pos {
	for _, a := range f.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// Column returns the first column that matches the given name.
func (f *ForeignKey) Column(name string) (*Column, bool) {
	for _, c := range f.Columns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// RefColumn returns the first referenced column that matches the given name.
func (f *ForeignKey) RefColumn(name string) (*Column, bool) {
	for _, c := range f.RefColumns {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// SetPos sets the position of the trigger.
func (t *Trigger) SetPos(p *Pos) {
	ReplaceOrAppend(&t.Attrs, p)
}

// Pos of the schema, if exists.
func (t *Trigger) Pos() *Pos {
	for _, a := range t.Attrs {
		if p, ok := a.(*Pos); ok {
			return p
		}
	}
	return nil
}

// ReferenceOption for constraint actions.
type ReferenceOption string

// Reference options (actions) specified by ON UPDATE and ON DELETE
// subclauses of the FOREIGN KEY clause.
const (
	NoAction   ReferenceOption = "NO ACTION"
	Restrict   ReferenceOption = "RESTRICT"
	Cascade    ReferenceOption = "CASCADE"
	SetNull    ReferenceOption = "SET NULL"
	SetDefault ReferenceOption = "SET DEFAULT"
)

// Common reference option constants for convenience.
var (
	OnDeleteCascade  = Cascade
	OnDeleteRestrict = Restrict
	OnDeleteNoAction = NoAction
	OnDeleteSetNull  = SetNull
	OnUpdateCascade  = Cascade
	OnUpdateRestrict = Restrict
	OnUpdateNoAction = NoAction
	OnUpdateSetNull  = SetNull
)

type (
	// A Type represents a database type. The types below implements this
	// interface and can be used for describing schemas.
	//
	// The Type interface can also be implemented outside this package as follows:
	//
	//	type SpatialType struct {
	//		schema.Type
	//		T string
	//	}
	//
	//	var t schema.Type = &SpatialType{T: "point"}
	//
	Type interface {
		typ()
	}

	EnumType struct {
		T      string
		Values []string
		Schema *Schema
		Attrs  []Attr
	}

	BinaryType struct {
		T    string
		Size *int
	}

	StringType struct {
		T     string
		Size  int
		Attrs []Attr
	}

	BoolType struct {
		T string
	}

	IntegerType struct {
		T        string
		Unsigned bool
		Attrs    []Attr
	}

	DecimalType struct {
		T         string
		Precision int
		Scale     int
		Unsigned  bool
	}

	FloatType struct {
		T         string
		Unsigned  bool
		Precision int
	}

	TimeType struct {
		T         string
		Precision *int
		Scale     *int
		Attrs     []Attr
	}

	JSONType struct {
		T string
	}

	SpatialType struct {
		T string
	}

	UUIDType struct {
		T string
	}

	UnsupportedType struct {
		T string
	}

	BitType struct {
		T   string
		Len int
	}

	TypeParser interface {
		ParseType(string) (Type, error)
	}

	TypeFormatter interface {
		FormatType(Type) (string, error)
	}

	TypeParseFormatter interface {
		TypeParser
		TypeFormatter
	}
)

type (
	// Expr defines an SQL expression in schema DDL.
	//
	// The Expr interface can also be implemented outside this package as follows:
	//
	// 	type NamedDefault struct {
	// 		schema.Expr
	// 		Name string
	// 	}
	// 	// Underlying returns the underlying expression.
	// 	func (e *NamedDefault) Underlying() schema.Expr { return e.Expr }
	//
	//  var e schema.Expr = &NamedDefault{Expr: &schema.Literal{V: "bar"}, Name: "foo"}
	Expr interface {
		expr()
	}

	Literal struct {
		V string
	}

	RawExpr struct {
		X string
	}
)

type (
	Attr interface {
		attr()
	}

	Comment struct {
		Text string
	}

	Charset struct {
		V string
	}

	Collation struct {
		V string
	}

	Check struct {
		Name  string
		Expr  string
		Attrs []Attr
	}

	GeneratedExpr struct {
		Expr string
		Type string
	}

	ViewCheckOption struct {
		V string
	}

	Materialized struct{}
)

// Implement the Attr interface for Materialized
func (*Materialized) attr() {}

type Pos struct {
	Filename   string
	Start, End struct {
		Line, Column, Byte int
	}
}

// String returns the position in editor/LSP style.
// Format: "filename:line[:c][-end_line[:end_c]]"
func (p *Pos) String() string {
	if p == nil {
		return ""
	}
	var b strings.Builder
	if p.Filename != "" {
		b.WriteString(p.Filename)
	} else {
		b.WriteByte('-')
	}
	if p.Start.Line > 0 {
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(p.Start.Line))
		if p.Start.Column > 0 {
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(p.Start.Column))
		}
	}
	return b.String()
}

// Underlying returns underlying the expression.
func (n *NamedDefault) Underlying() Expr {
	return n.Expr
}

// UnderlyingExpr returns the underlying expression of x.
func UnderlyingExpr(x Expr) Expr {
	if w, ok := x.(interface{ Underlying() Expr }); ok {
		return UnderlyingExpr(w.Underlying())
	}
	return x
}

// UnderlyingType returns the underlying type of t.
func UnderlyingType(t Type) Type {
	if w, ok := t.(interface{ Underlying() Type }); ok {
		return UnderlyingType(w.Underlying())
	}
	return t
}

// IsType returns true if somewhere in the type-chain of t1 is the same as t2.
func IsType(t1, t2 Type) bool {
	if t1 == nil || t2 == nil {
		return t1 == t2
	}
	return sameType(t1, t2, reflect.TypeOf(t2).Comparable())
}

// sameType is a helper function for IsType.
func sameType(t1, t2 Type, targetComparable bool) bool {
	for {
		if targetComparable && t1 == t2 {
			return true
		}
		if x, ok := t1.(interface{ Is(Type) bool }); ok && x.Is(t2) {
			return true
		}
		if x, ok := t1.(interface{ Underlying() Type }); ok {
			if t1 = x.Underlying(); t1 != nil {
				continue
			}
		}
		return false
	}
}

type Option func(interface{})
type option = Option

type CascadeOn struct{}

func (*CascadeOn) attr() {}
