// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

// Attr interface implementations for all attribute types.
func (c *Check) attr()           {}
func (g *GeneratedExpr) attr()   {}
func (v *ViewCheckOption) attr() {}
func (m *Materialized) attr()    {}
func (c *Charset) attr()         {}
func (c *Collation) attr()       {}
func (c *Comment) attr()         {}
func (c *CascadeOn) attr()       {}

// ReplaceOrAppend replaces an existing attribute of the same type in the slice
// or appends the new attribute if none exists.
func ReplaceOrAppend(attrs *[]Attr, attr Attr) {
	t := Name(attr)
	for i, a := range *attrs {
		if Name(a) == t {
			(*attrs)[i] = attr
			return
		}
	}
	*attrs = append(*attrs, attr)
}

// RemoveAttr removes an attribute of type T from the slice.
func RemoveAttr[T any](attrs []Attr) []Attr {
	var name string
	switch any(interface{}(new(T))).(type) {
	case **Charset, *Charset:
		name = "charset"
	case **Collation, *Collation:
		name = "collation"
	case **Comment, *Comment:
		name = "comment"
	default:
		// Cannot determine the type name
		return attrs
	}

	for i, a := range attrs {
		if Name(a) == name {
			return append(attrs[:i], attrs[i+1:]...)
		}
	}
	return attrs
}

// del deletes an attribute of type T from the slice.
func del(attrs []Attr, typ Attr) []Attr {
	targetName := Name(typ)
	for i, a := range attrs {
		if Name(a) == targetName {
			return append(attrs[:i], attrs[i+1:]...)
		}
	}
	return attrs
}
