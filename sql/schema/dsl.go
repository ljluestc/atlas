// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"reflect"
	"slices"
	"strings"
)

// The functions and methods below provide a DSL for creating schema resources using
// a fluent interface. Note that some methods create links between the schema elements.

// New creates a new Schema.
func New(name string) *Schema {
	return &Schema{Name: name}
}

// SetCharset sets or appends the Charset attribute
// to the schema with the given value.
func (s *Schema) SetCharset(v string) *Schema {
	ReplaceOrAppend(&s.Attrs, &Charset{V: v})
	return s
}

// UnsetCharset unsets the Charset attribute.
func (s *Schema) UnsetCharset() *Schema {
	s.Attrs = RemoveAttr[*Charset](s.Attrs)
	return s
}

// SetCollation sets or appends the Collation attribute
// to the schema with the given value.
func (s *Schema) SetCollation(v string) *Schema {
	ReplaceOrAppend(&s.Attrs, &Collation{V: v})
	return s
}

// UnsetCollation unsets the Collation attribute.
func (s *Schema) UnsetCollation() *Schema {
	s.Attrs = RemoveAttr[*Collation](s.Attrs)
	return s
}

// SetComment sets or appends the Comment attribute
// to the schema with the given value.
func (s *Schema) SetComment(v string) *Schema {
	ReplaceOrAppend(&s.Attrs, &Comment{Text: v})
	return s
}

// AddAttrs adds additional attributes to the schema.
func (s *Schema) AddAttrs(attrs ...Attr) *Schema {
	s.Attrs = append(s.Attrs, attrs...)
	return s
}

// SetRealm sets the database/realm of the schema.
// This is a no-op as Schema has no Realm field.
func (s *Schema) SetRealm(r *Realm) *Schema {
	// No Realm field in Schema, so this is a no-op
	return s
}

// AddTables adds and links the given tables to the schema.
func (s *Schema) AddTables(tables ...*Table) *Schema {
	for _, t := range tables {
		t.SetSchema(s)
	}
	if s.Tables == nil {
		s.Tables = tables
	} else {
		s.Tables = append(s.Tables, tables...)
	}
	return s
}

// AddViews adds and links the given views to the schema.
func (s *Schema) AddViews(views ...*View) *Schema {
	for _, v := range views {
		v.SetSchema(s)
	}
	if s.Views == nil {
		s.Views = views
	} else {
		temp := append(s.Views, views...)
		s.Views = temp
	}
	return s
}

// AddObjects adds the given objects to the schema.
func (s *Schema) AddObjects(objs ...Object) *Schema {
	if s.Objects == nil {
		s.Objects = objs
	} else {
		s.Objects = append(s.Objects, objs...)
	}
	return s
}

// AddFuncs appends the given functions to the schema.
func (s *Schema) AddFuncs(funcs ...*Func) *Schema {
	for _, f := range funcs {
		f.Schema = s
	}
	s.Funcs = append(s.Funcs, funcs...)
	return s
}

// AddProcs appends the given procedures to the schema.
func (s *Schema) AddProcs(procs ...*Proc) *Schema {
	for _, f := range procs {
		f.Schema = s
	}
	s.Procs = append(s.Procs, procs...)
	return s
}

// NewRealm creates a new Realm.
func NewRealm(schemas ...*Schema) *Realm {
	r := &Realm{Schemas: schemas}
	// No Realm field in Schema, so this is a no-op
	// for _, s := range schemas {
	//     s.Realm = r
	// }
	return r
}

// AddSchemas adds and links the given schemas to the realm.
func (r *Realm) AddSchemas(schemas ...*Schema) *Realm {
	for _, s := range schemas {
		s.SetRealm(r)
	}
	r.Schemas = append(r.Schemas, schemas...)
	return r
}

// AddObjects adds the given objects to the realm.
func (r *Realm) AddObjects(objs ...Object) *Realm {
	if r.Objects == nil {
		r.Objects = objs
	} else {
		r.Objects = append(r.Objects, objs...)
	}
	return r
}

// SetCharset sets or appends the Charset attribute
// to the realm with the given value.
func (r *Realm) SetCharset(v string) *Realm {
	ReplaceOrAppend(&r.Attrs, &Charset{V: v})
	return r
}

// UnsetCharset unsets the Charset attribute.
func (r *Realm) UnsetCharset() *Realm {
	r.Attrs = RemoveAttr[*Charset](r.Attrs)
	return r
}

// SetCollation sets or appends the Collation attribute
// to the realm with the given value.
func (r *Realm) SetCollation(v string) *Realm {
	ReplaceOrAppend(&r.Attrs, &Collation{V: v})
	return r
}

// UnsetCollation unsets the Collation attribute.
func (r *Realm) UnsetCollation() *Realm {
	r.Attrs = RemoveAttr[*Collation](r.Attrs)
	return r
}

// SetComment sets or appends the Comment attribute
// to the realm with the given value.
func (r *Realm) SetComment(v string) *Realm {
	ReplaceOrAppend(&r.Attrs, &Comment{Text: v})
	return r
}

// NewTable creates a new Table.
func NewTable(name string) *Table {
	return &Table{Name: name}
}

// SetCharset sets or appends the Charset attribute
// to the table with the given value.
func (t *Table) SetCharset(v string) *Table {
	ReplaceOrAppend(&t.Attrs, &Charset{V: v})
	return t
}

// UnsetCharset unsets the Charset attribute.
func (t *Table) UnsetCharset() *Table {
	t.Attrs = RemoveAttr[*Charset](t.Attrs)
	return t
}

// SetCollation sets or appends the Collation attribute
// to the table with the given value.
func (t *Table) SetCollation(v string) *Table {
	ReplaceOrAppend(&t.Attrs, &Collation{V: v})
	return t
}

// UnsetCollation unsets the Collation attribute.
func (t *Table) UnsetCollation() *Table {
	t.Attrs = RemoveAttr[*Collation](t.Attrs)
	return t
}

// SetComment sets or appends the Comment attribute
// to the table with the given value.
func (t *Table) SetComment(v string) *Table {
	ReplaceOrAppend(&t.Attrs, &Comment{Text: v})
	return t
}

// AddChecks appends the given checks to the attribute list.
func (t *Table) AddChecks(checks ...*Check) *Table {
	for _, c := range checks {
		t.Attrs = append(t.Attrs, c)
	}
	return t
}

// SetSchema sets the schema (named-database) of the table.
func (t *Table) SetSchema(s *Schema) *Table {
	if s != nil {
		t.Schema = s
		s.AddTables(t)
	}
	return t
}

// SetPrimaryKey sets the primary-key of the table.
func (t *Table) SetPrimaryKey(pk *Index) *Table {
	pk.Table = t
	t.PrimaryKey = pk
	for _, p := range pk.Parts {
		if p.C == nil {
			continue
		}
		if _, ok := t.Column(p.C.Name); !ok {
			t.AddColumns(p.C)
		}
	}
	return t
}

// AddColumns appends the given columns to the table column list.
func (t *Table) AddColumns(columns ...*Column) *Table {
	t.Columns = append(t.Columns, columns...)
	return t
}

// AddIndexes appends the given indexes to the table index list.
func (t *Table) AddIndexes(indexes ...*Index) *Table {
	for _, idx := range indexes {
		idx.Table = t
	}
	t.Indexes = append(t.Indexes, indexes...)
	return t
}

// AddForeignKeys appends the given foreign-keys to the table foreign-key list.
func (t *Table) AddForeignKeys(fks ...*ForeignKey) *Table {
	for _, fk := range fks {
		fk.Table = t
	}
	t.ForeignKeys = append(t.ForeignKeys, fks...)
	return t
}

// AddAttrs adds and additional attributes to the table.
func (t *Table) AddAttrs(attrs ...Attr) *Table {
	t.Attrs = append(t.Attrs, attrs...)
	return t
}

// AddDeps adds the given objects as dependencies to the view.
func (t *Table) AddDeps(objs ...Object) *Table {
	t.Deps = append(t.Deps, objs...)
	addRefs(t, objs)
	return t
}

// RefsAdder wraps the AddRefs method. Objects that implemented this method
// will get their dependent objects automatically set by their AddDeps calls.
type RefsAdder interface {
	AddRefs(...Object)
}

// addRefs adds the dependent objects to all objects it references.
func addRefs(dependent Object, refs []Object) {
	for _, o := range refs {
		if r, ok := o.(RefsAdder); ok {
			r.AddRefs(dependent)
		}
	}
}

// DepRemover wraps the RemoveDep method. Objects that implemented this method
// allow dependent objects to remove themselves from their references.
type DepRemover interface {
	RemoveDep(Object)
}

// removeObj removes the given object from the list of objects.
func removeObj(objs []Object, o Object) []Object {
	i := slices.Index(objs, o)
	if i == -1 {
		return objs
	}
	return append(objs[:i:i], objs[i+1:]...)
}

// SortRefs maintains consistent dependents list.
func SortRefs(refs []Object) {
	slices.SortFunc(refs, func(a, b Object) int {
		typeA, typeB := reflect.TypeOf(a), reflect.TypeOf(b)
		if typeA != typeB {
			return strings.Compare(typeA.String(), typeB.String())
		}
		switch o1 := a.(type) {
		case *Table:
			return strings.Compare(o1.Name, b.(*Table).Name)
		case *View:
			return strings.Compare(o1.Name, b.(*View).Name)
		case *Trigger:
			return strings.Compare(o1.Name, b.(*Trigger).Name)
		case *Func:
			return strings.Compare(o1.Name, b.(*Func).Name)
		case *Proc:
			return strings.Compare(o1.Name, b.(*Proc).Name)
		default:
			return 0
		}
	})
}

// AddRefs adds references to the table.
func (t *Table) AddRefs(refs ...Object) {
	t.Refs = append(t.Refs, refs...)
	SortRefs(t.Refs)
}

// RemoveDep removes the given object from the table dependencies.
func (t *Table) RemoveDep(o Object) {
	t.Deps = removeObj(t.Deps, o)
}

// NewView creates a new View.
func NewView(name, def string) *View {
	return &View{Name: name, Def: def}
}

// NewMaterializedView creates a new materialized View.
func NewMaterializedView(name, def string) *View {
	return NewView(name, def).
		SetMaterialized(true)
}

// SetSchema sets the schema (named-database) of the view.
func (v *View) SetSchema(s *Schema) *View {
	v.Schema = s
	return v
}

// AddColumns appends the given columns to the table column list.
func (v *View) AddColumns(columns ...*Column) *View {
	v.Columns = append(v.Columns, columns...)
	return v
}

// SetComment sets or appends the Comment attribute
// to the view with the given value.
func (v *View) SetComment(c string) *View {
	ReplaceOrAppend(&v.Attrs, &Comment{Text: c})
	return v
}

// AddAttrs adds and additional attributes to the view.
func (v *View) AddAttrs(attrs ...Attr) *View {
	v.Attrs = append(v.Attrs, attrs...)
	return v
}

// AddDeps adds the given objects as dependencies to the view.
func (v *View) AddDeps(objs ...Object) *View {
	v.Deps = append(v.Deps, objs...)
	addRefs(v, objs)
	return v
}

// RemoveDep removes the given object from the view dependencies.
func (v *View) RemoveDep(o Object) {
	v.Deps = removeObj(v.Deps, o)
}

// AddRefs adds references to the view.
func (v *View) AddRefs(refs ...Object) {
	v.Refs = append(v.Refs, refs...)
	SortRefs(v.Refs)
}

// AddIndexes appends the given indexes to the table index list.
func (v *View) AddIndexes(indexes ...*Index) *View {
	for _, idx := range indexes {
		idx.View = v
	}
	v.Indexes = append(v.Indexes, indexes...)
	return v
}

// SetCheckOption sets the check option of the view.
func (v *View) SetCheckOption(opt string) *View {
	ReplaceOrAppend(&v.Attrs, &ViewCheckOption{V: opt})
	return v
}

// NewTrigger creates a new Trigger with the given name.
func NewTrigger(name string) *Trigger {
	return &Trigger{Name: name}
}

// SetSchema sets the schema of the trigger.
func (t *Trigger) SetSchema(s *Schema) *Trigger {
	t.Schema = s
	return t
}

// SetTable sets the table of the trigger.
func (t *Trigger) SetTable(table *Table) *Trigger {
	t.Table = table
	return t
}

// SetView sets the view of the trigger.
func (t *Trigger) SetView(view *View) *Trigger {
	t.View = view
	return t
}

// SetTime sets the time when the trigger executes (BEFORE, AFTER, etc).
func (t *Trigger) SetTime(time string) *Trigger {
	t.Time = time
	return t
}

// SetEvent sets the event that activates the trigger (INSERT, UPDATE, DELETE).
func (t *Trigger) SetEvent(event string) *Trigger {
	t.Event = event
	return t
}

// SetEvents sets multiple events that activate the trigger.
func (t *Trigger) SetEvents(events ...string) *Trigger {
	t.Events = events
	return t
}

// SetStatement sets the SQL statement executed by the trigger.
func (t *Trigger) SetStatement(stmt string) *Trigger {
	t.Statement = stmt
	return t
}

// SetComment sets or appends the Comment attribute to the trigger with the given value.
func (t *Trigger) SetComment(v string) *Trigger {
	ReplaceOrAppend(&t.Attrs, &Comment{Text: v})
	return t
}

// AddAttrs adds additional attributes to the trigger.
func (t *Trigger) AddAttrs(attrs ...Attr) *Trigger {
	t.Attrs = append(t.Attrs, attrs...)
	return t
}

// AddDeps adds the given objects as dependencies to the trigger.
func (t *Trigger) AddDeps(objs ...Object) *Trigger {
	t.Deps = append(t.Deps, objs...)
	addRefs(t, objs)
	return t
}

// RemoveDep removes the given object from the trigger dependencies.
func (t *Trigger) RemoveDep(o Object) {
	t.Deps = removeObj(t.Deps, o)
}

// AddRefs adds references to the trigger.
func (t *Trigger) AddRefs(refs ...Object) {
	t.Refs = append(t.Refs, refs...)
	SortRefs(t.Refs)
}

// NewColumn creates a new column with the given name.
func NewColumn(name string) *Column {
	return &Column{Name: name}
}

// NewNullColumn creates a new nullable column with the given name.
func NewNullColumn(name string) *Column {
	return NewColumn(name).
		SetNull(true)
}

// NewBoolColumn creates a new BoolType column.
func NewBoolColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&BoolType{T: typ})
}

// NewNullBoolColumn creates a new nullable BoolType column.
func NewNullBoolColumn(name, typ string) *Column {
	return NewBoolColumn(name, typ).
		SetNull(true)
}

// NewIntColumn creates a new IntegerType column.
func NewIntColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&IntegerType{T: typ})
}

// NewNullIntColumn creates a new nullable IntegerType column.
func NewNullIntColumn(name, typ string) *Column {
	return NewIntColumn(name, typ).
		SetNull(true)
}

// NewUintColumn creates a new unsigned IntegerType column.
func NewUintColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&IntegerType{T: typ, Unsigned: true})
}

// NewNullUintColumn creates a new nullable unsigned IntegerType column.
func NewNullUintColumn(name, typ string) *Column {
	return NewUintColumn(name, typ).
		SetNull(true)
}

// EnumOption allows configuring EnumType using functional options.
type EnumOption func(*EnumType)

// EnumName configures the name of the name. This option
// is useful for databases like PostgreSQL that supports
// user-defined types for enums.
func EnumName(name string) EnumOption {
	return func(e *EnumType) {
		e.T = name
	}
}

// EnumValues configures the values of the enum.
func EnumValues(values ...string) EnumOption {
	return func(e *EnumType) {
		e.Values = values
	}
}

// EnumSchema configures the schema of the enum.
func EnumSchema(s *Schema) EnumOption {
	return func(e *EnumType) {
		e.Schema = s
	}
}

// NewEnumColumn creates a new EnumType column.
func NewEnumColumn(name string, opts ...EnumOption) *Column {
	t := &EnumType{}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

// NewNullEnumColumn creates a new nullable EnumType column.
func NewNullEnumColumn(name string, opts ...EnumOption) *Column {
	return NewEnumColumn(name, opts...).
		SetNull(true)
}

// BinaryOption allows configuring BinaryType using functional options.
type BinaryOption func(*BinaryType)

// BinarySize configures the size of the binary type.
func BinarySize(size int) BinaryOption {
	return func(b *BinaryType) {
		b.Size = &size
	}
}

// NewBinaryColumn creates a new BinaryType column.
func NewBinaryColumn(name, typ string, opts ...BinaryOption) *Column {
	t := &BinaryType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

// NewNullBinaryColumn creates a new nullable BinaryType column.
func NewNullBinaryColumn(name, typ string, opts ...BinaryOption) *Column {
	return NewBinaryColumn(name, typ, opts...).
		SetNull(true)
}

// StringOption allows configuring StringType using functional options.
type StringOption func(*StringType)

// StringSize configures the size of the string type.
func StringSize(size int) StringOption {
	return func(b *StringType) {
		b.Size = size
	}
}

// NewStringColumn creates a new StringType column.
func NewStringColumn(name, typ string, opts ...StringOption) *Column {
	t := &StringType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

// NewNullStringColumn creates a new nullable StringType column.
func NewNullStringColumn(name, typ string, opts ...StringOption) *Column {
	return NewStringColumn(name, typ, opts...).
		SetNull(true)
}

// DecimalOption allows configuring DecimalType using functional options.
type DecimalOption func(*DecimalType)

// DecimalPrecision configures the precision of the decimal type.
func DecimalPrecision(precision int) DecimalOption {
	return func(b *DecimalType) {
		b.Precision = precision
	}
}

// DecimalScale configures the scale of the decimal type.
func DecimalScale(scale int) DecimalOption {
	return func(b *DecimalType) {
		b.Scale = scale
	}
}

// DecimalUnsigned configures the unsigned of the float type.
func DecimalUnsigned(unsigned bool) DecimalOption {
	return func(b *DecimalType) {
		b.Unsigned = unsigned
	}
}

// NewDecimalColumn creates a new DecimalType column.
func NewDecimalColumn(name, typ string, opts ...DecimalOption) *Column {
	t := &DecimalType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

// NewNullDecimalColumn creates a new nullable DecimalType column.
func NewNullDecimalColumn(name, typ string, opts ...DecimalOption) *Column {
	return NewDecimalColumn(name, typ, opts...).
		SetNull(true)
}

// FloatOption allows configuring FloatType using functional options.
type FloatOption func(*FloatType)

// FloatPrecision configures the precision of the float type.
func FloatPrecision(precision int) FloatOption {
	return func(b *FloatType) {
		b.Precision = precision
	}
}

// FloatUnsigned configures the unsigned of the float type.
func FloatUnsigned(unsigned bool) FloatOption {
	return func(b *FloatType) {
		b.Unsigned = unsigned
	}
}

// NewFloatColumn creates a new FloatType column.
func NewFloatColumn(name, typ string, opts ...FloatOption) *Column {
	t := &FloatType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

// NewNullFloatColumn creates a new nullable FloatType column.
func NewNullFloatColumn(name, typ string, opts ...FloatOption) *Column {
	return NewFloatColumn(name, typ, opts...).
		SetNull(true)
}

// TimeOption allows configuring TimeType using functional options.
type TimeOption func(*TimeType)

// TimePrecision configures the precision of the time type.
func TimePrecision(precision int) TimeOption {
	return func(b *TimeType) {
		b.Precision = &precision
	}
}

// TimeScale configures the scale of the time type.
func TimeScale(scale int) TimeOption {
	return func(b *TimeType) {
		b.Scale = &scale
	}
}

// NewTimeColumn creates a new TimeType column.
func NewTimeColumn(name, typ string, opts ...TimeOption) *Column {
	t := &TimeType{T: typ}
	for _, opt := range opts {
		opt(t)
	}
	return NewColumn(name).SetType(t)
}

// NewNullTimeColumn creates a new nullable TimeType column.
func NewNullTimeColumn(name, typ string) *Column {
	return NewTimeColumn(name, typ).
		SetNull(true)
}

// NewJSONColumn creates a new JSONType column.
func NewJSONColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&JSONType{T: typ})
}

// NewNullJSONColumn creates a new nullable JSONType column.
func NewNullJSONColumn(name, typ string) *Column {
	return NewJSONColumn(name, typ).
		SetNull(true)
}

// NewSpatialColumn creates a new SpatialType column.
func NewSpatialColumn(name, typ string) *Column {
	return NewColumn(name).
		SetType(&SpatialType{T: typ})
}

// NewNullSpatialColumn creates a new nullable SpatialType column.
func NewNullSpatialColumn(name, typ string) *Column {
	return NewSpatialColumn(name, typ).
		SetNull(true)
}

// SetNull configures the nullability of the column
func (c *Column) SetNull(b bool) *Column {
	if c.Type == nil {
		c.Type = &ColumnType{}
	}
	c.Type.Null = b
	return c
}

// SetType configures the type of the column
func (c *Column) SetType(t Type) *Column {
	if c.Type == nil {
		c.Type = &ColumnType{}
	}
	c.Type.Type = t
	return c
}

// SetDefault configures the default of the column
func (c *Column) SetDefault(x Expr) *Column {
	c.Default = x
	return c
}

// SetCharset sets or appends the Charset attribute
// to the column with the given value.
func (c *Column) SetCharset(v string) *Column {
	ReplaceOrAppend(&c.Attrs, &Charset{V: v})
	return c
}

// UnsetCharset unsets the Charset attribute.
func (c *Column) UnsetCharset() *Column {
	c.Attrs = RemoveAttr[*Charset](c.Attrs)
	return c
}

// SetCollation sets or appends the Collation attribute
// to the column with the given value.
func (c *Column) SetCollation(v string) *Column {
	ReplaceOrAppend(&c.Attrs, &Collation{V: v})
	return c
}

// UnsetCollation unsets the Collation attribute.
func (c *Column) UnsetCollation() *Column {
	c.Attrs = RemoveAttr[*Collation](c.Attrs)
	return c
}

// SetComment sets or appends the Comment attribute
// to the column with the given value.
func (c *Column) SetComment(v string) *Column {
	ReplaceOrAppend(&c.Attrs, &Comment{Text: v})
	return c
}

// SetGeneratedExpr sets or appends the GeneratedExpr attribute.
func (c *Column) SetGeneratedExpr(x *GeneratedExpr) *Column {
	ReplaceOrAppend(&c.Attrs, x)
	return c
}

// AddAttrs adds additional attributes to the column.
func (c *Column) AddAttrs(attrs ...Attr) *Column {
	c.Attrs = append(c.Attrs, attrs...)
	return c
}

// AddIndexes appends the references to the indexes this column is part of.
func (c *Column) AddIndexes(indexes ...*Index) *Column {
	for _, idx := range indexes {
		if !slices.Contains(c.Indexes, idx) {
			c.Indexes = append(c.Indexes, idx)
		}
	}
	return c
}

// NewCheck creates a new check.
func NewCheck() *Check {
	return &Check{}
}

// SetName configures the name of the check constraint.
func (c *Check) SetName(name string) *Check {
	c.Name = name
	return c
}

// SetExpr configures the expression of the check constraint.
func (c *Check) SetExpr(expr string) *Check {
	c.Expr = expr
	return c
}

// AddAttrs adds additional attributes to the check constraint.
func (c *Check) AddAttrs(attrs ...Attr) *Check {
	c.Attrs = append(c.Attrs, attrs...)
	return c
}

// NewIndex creates a new index with the given name.
func NewIndex(name string) *Index {
	return &Index{Name: name}
}

// NewUniqueIndex creates a new unique index with the given name.
func NewUniqueIndex(name string) *Index {
	return NewIndex(name).SetUnique(true)
}

// NewPrimaryKey creates a new primary-key index
// for the given columns.
func NewPrimaryKey(columns ...*Column) *Index {
	return new(Index).SetUnique(true).AddColumns(columns...)
}

// SetName configures the name of the index.
func (i *Index) SetName(name string) *Index {
	i.Name = name
	return i
}

// SetUnique configures the uniqueness of the index.
func (i *Index) SetUnique(b bool) *Index {
	i.Unique = b
	return i
}

// SetTable configures the table of the index.
func (i *Index) SetTable(t *Table) *Index {
	i.Table = t
	return i
}

// SetComment sets or appends the Comment attribute
// to the index with the given value.
func (i *Index) SetComment(v string) *Index {
	ReplaceOrAppend(&i.Attrs, &Comment{Text: v})
	return i
}

// AddAttrs adds additional attributes to the index.
func (i *Index) AddAttrs(attrs ...Attr) *Index {
	i.Attrs = append(i.Attrs, attrs...)
	return i
}

// AddColumns adds the columns to index parts.
func (i *Index) AddColumns(columns ...*Column) *Index {
	for _, c := range columns {
		if !c.hasIndex(i) {
			c.Indexes = append(c.Indexes, i)
		}
		i.Parts = append(i.Parts, &IndexPart{SeqNo: len(i.Parts), C: c})
	}
	return i
}

func (c *Column) hasIndex(idx *Index) bool {
	for i := range c.Indexes {
		if c.Indexes[i] == idx {
			return true
		}
	}
	return false
}

// AddExprs adds the expressions to index parts.
func (i *Index) AddExprs(exprs ...Expr) *Index {
	for _, x := range exprs {
		i.Parts = append(i.Parts, &IndexPart{SeqNo: len(i.Parts), X: x})
	}
	return i
}

// AddParts appends the given parts.
func (i *Index) AddParts(parts ...*IndexPart) *Index {
	for _, p := range parts {
		if p.C != nil && !p.C.hasIndex(i) {
			p.C.Indexes = append(p.C.Indexes, i)
		}
		p.SeqNo = len(i.Parts)
		i.Parts = append(i.Parts, p)
	}
	return i
}

// NewIndexPart creates a new index part.
func NewIndexPart() *IndexPart { return &IndexPart{} }

// NewColumnPart creates a new index part with the given column.
func NewColumnPart(c *Column) *IndexPart { return NewIndexPart().SetColumn(c) }

// NewExprPart creates a new index part with the given expression.
func NewExprPart(x Expr) *IndexPart { return NewIndexPart().SetExpr(x) }

// SetDesc configures the "DESC" attribute of the key part.
func (p *IndexPart) SetDesc(b bool) *IndexPart {
	p.Desc = b
	return p
}

// AddAttrs adds and additional attributes to the index-part.
func (p *IndexPart) AddAttrs(attrs ...Attr) *IndexPart {
	p.Attrs = append(p.Attrs, attrs...)
	return p
}

// SetColumn sets the column of the index-part.
func (p *IndexPart) SetColumn(c *Column) *IndexPart {
	p.C = c
	return p
}

// SetExpr sets the expression of the index-part.
func (p *IndexPart) SetExpr(x Expr) *IndexPart {
	p.X = x
	return p
}

// NewForeignKey creates a new foreign-key with
// the given constraints/symbol name.
func NewForeignKey(symbol string) *ForeignKey {
	return &ForeignKey{Symbol: symbol}
}

// SetTable configures the table that holds the foreign-key (child table).
func (f *ForeignKey) SetTable(t *Table) *ForeignKey {
	f.Table = t
	return f
}

// AddColumns appends columns to the child-table columns.
func (f *ForeignKey) AddColumns(columns ...*Column) *ForeignKey {
	for _, c := range columns {
		if !c.hasForeignKey(f) {
			c.ForeignKeys = append(c.ForeignKeys, f)
		}
	}
	f.Columns = append(f.Columns, columns...)
	return f
}

func (c *Column) hasForeignKey(fk *ForeignKey) bool {
	for i := range c.ForeignKeys {
		if c.ForeignKeys[i] == fk {
			return true
		}
	}
	return false
}

// SetRefTable configures the referenced/parent table.
func (f *ForeignKey) SetRefTable(t *Table) *ForeignKey {
	f.RefTable = t
	return f
}

// AddRefColumns appends columns to the parent-table columns.
func (f *ForeignKey) AddRefColumns(columns ...*Column) *ForeignKey {
	f.RefColumns = append(f.RefColumns, columns...)
	return f
}

// SetOnUpdate sets the ON UPDATE constraint action.
func (f *ForeignKey) SetOnUpdate(o ReferenceOption) *ForeignKey {
	f.OnUpdate = o
	return f
}

// SetOnDelete sets the ON DELETE constraint action.
func (f *ForeignKey) SetOnDelete(o ReferenceOption) *ForeignKey {
	f.OnDelete = o
	return f
}

// AddAttrs adds additional attributes to the schema.
func (f *ForeignKey) AddAttrs(attrs ...Attr) *ForeignKey {
	f.Attrs = append(f.Attrs, attrs...)
	return f
}

// AddDeps adds the given objects as dependencies to the function.
func (f *Func) AddDeps(objs ...Object) *Func {
	f.Deps = append(f.Deps, objs...)
	addRefs(f, objs)
	return f
}

// RemoveDep removes the given object from the function dependencies.
func (f *Func) RemoveDep(o Object) {
	f.Deps = removeObj(f.Deps, o)
}

// AddRefs adds references to the function.
func (f *Func) AddRefs(refs ...Object) {
	f.Refs = append(f.Refs, refs...)
	SortRefs(f.Refs)
}

// AddDeps adds the given objects as dependencies to the procedure.
func (p *Proc) AddDeps(objs ...Object) *Proc {
	p.Deps = append(p.Deps, objs...)
	addRefs(p, objs)
	return p
}

// RemoveDep removes the given object from the procedure dependencies.
func (p *Proc) RemoveDep(o Object) {
	p.Deps = removeObj(p.Deps, o)
}

// AddRefs adds references to the procedure.
func (p *Proc) AddRefs(refs ...Object) {
	p.Refs = append(p.Refs, refs...)
	SortRefs(p.Refs)
}

// SetMaterialized sets the materialized attribute of the view.
func (v *View) SetMaterialized(b bool) *View {
	ReplaceOrAppend(&v.Attrs, &Materialized{V: b})
	return v
}

// ColumnType needs to implement Type interface
func (c *ColumnType) typ() {}

// Trigger implementation for obj
func (t *Trigger) obj() {}

// Helper function to get name of an Attr
func Name(a Attr) string {
	switch a := a.(type) {
	case *Charset:
		return "charset"
	case *Collation:
		return "collation"
	case *Comment:
		return "comment"
	default:
		// Default type name
		return reflect.TypeOf(a).String()
	}
}
