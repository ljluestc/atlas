// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

// ReplaceOrAppend replaces the first attribute of type T in the slice with the given attribute.
// If no such attribute exists, it appends the given attribute to the slice.
func ReplaceOrAppend[T Attr, P *T](attrs *[]Attr, attr P) {
	for i, a := range *attrs {
		if _, ok := a.(T); ok {
			(*attrs)[i] = attr
			return
		}
	}
	*attrs = append(*attrs, attr)
}

// RemoveAttr removes all attributes of type T from the given slice.
func RemoveAttr[T Attr](attrs []Attr) []Attr {
	var filtered []Attr
	for _, a := range attrs {
		if _, ok := a.(T); !ok {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// Is reports if the given type is the same as this type.
func (t *EnumType) Is(other Type) bool {
	o, ok := other.(*EnumType)
	if !ok {
		return false
	}
	if t.T != o.T {
		return false
	}
	if len(t.Values) != len(o.Values) {
		return false
	}
	for i := range t.Values {
		if t.Values[i] != o.Values[i] {
			return false
		}
	}
	return true
}

// Interface implementations for interfaces.
func (*Schema) obj() {}
func (*Table) obj()  {}
func (*View) obj()   {}
func (*Func) obj()   {}
func (*Proc) obj()   {}

// Type interface implementations for all types.
func (*EnumType) typ()        {}
func (*BinaryType) typ()      {}
func (*StringType) typ()      {}
func (*BoolType) typ()        {}
func (*IntegerType) typ()     {}
func (*DecimalType) typ()     {}
func (*FloatType) typ()       {}
func (*TimeType) typ()        {}
func (*JSONType) typ()        {}
func (*SpatialType) typ()     {}
func (*UUIDType) typ()        {}
func (*UnsupportedType) typ() {}
func (*BitType) typ()         {}

// Expr interface implementations.
func (*Literal) expr() {}
func (*RawExpr) expr() {}

// Attr interface implementations.
// func (*Comment) attr()         {}
// func (*Charset) attr()         {}
// func (*Collation) attr()       {}
// func (*Check) attr()           {}
// func (*GeneratedExpr) attr()   {}
// func (*ViewCheckOption) attr() {}
func (*Pos) attr() {}

// SameType reports whether a1 and a2 are attributes of the same type.
func SameType(a1, a2 Attr) bool {
	switch a1.(type) {
	case *Comment:
		_, ok := a2.(*Comment)
		return ok
	case *Charset:
		_, ok := a2.(*Charset)
		return ok
	case *Collation:
		_, ok := a2.(*Collation)
		return ok
	case *Check:
		_, ok := a2.(*Check)
		return ok
	case *GeneratedExpr:
		_, ok := a2.(*GeneratedExpr)
		return ok
	case *ViewCheckOption:
		_, ok := a2.(*ViewCheckOption)
		return ok
	case *Materialized:
		_, ok := a2.(*Materialized)
		return ok
	case *Pos:
		_, ok := a2.(*Pos)
		return ok
	case *ExcludeConstraint:
		_, ok := a2.(*ExcludeConstraint)
		return ok
	case *CascadeOn:
		_, ok := a2.(*CascadeOn)
		return ok
	default:
		return false
	}
}

// The following functions are declared in dsl.go and attr.go, so we remove them here
// to avoid duplicate declarations.
