package object

import "fmt"

type ObjectType string

const (
	INTEGER = "INTEGER"
	BOOL    = "BOOL"
	NULL    = "NULL"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER }

type Bool struct {
	Value bool
}

func (b *Bool) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Bool) Type() ObjectType { return BOOL }

type Null struct {
}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL }
