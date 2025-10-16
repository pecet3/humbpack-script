package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"

	"github.com/pecet3/hmbk-script/ast"
)

type ObjectType string

const (
	NUMBER       = "NUMBER"
	INTEGER      = "INTEGER"
	BOOL         = "BOOL"
	STRING       = "STRING"
	NULL         = "NULL"
	RETURN_VALUE = "RETURN_VALUE"
	ERROR        = "ERROR"
	FUNCTION     = "FUNCTION"
	BUILTIN      = "BUILTIN"
	ARRAY        = "ARRAY"
	HASH         = "HASH"
	MODULE       = "MODULE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Number struct {
	Value float64
}

func (i *Number) Inspect() string {
	return strconv.FormatFloat(i.Value, 'f', -1, 64)
}
func (i *Number) Type() ObjectType { return NUMBER }
func (i *Number) Int() int64 {
	return int64(math.Round(i.Value))
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}
func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) Float() float64 {
	return float64(i.Value)
}

type Bool struct {
	Value bool
}

func (b *Bool) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Bool) Type() ObjectType { return BOOL }

type Null struct {
}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type Module struct {
	Name string
	Env  *Environment
}

func (f *Module) Type() ObjectType { return MODULE }
func (f *Module) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Env.consts {
		params = append(params, p.Inspect())
	}
	out.WriteString("module ")
	out.WriteString("{")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString("}\n")
	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }

type BuiltinFunction func(args ...Object) Object
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (b *Bool) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Number) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()

	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}

	for _, p := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", p.Key.Inspect(), p.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type Hashable interface {
	HashKey() HashKey
}
