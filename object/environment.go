package object

type Environment struct {
	store  map[string]Object
	consts map[string]Object
	outer  *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok {
		obj, ok = e.consts[name]
	}
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) IsConst(name string) bool {
	_, ok := e.consts[name]
	return ok
}
func (e *Environment) SetConst(name string, val Object) Object {
	e.consts[name] = val
	return val
}
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	c := make(map[string]Object)
	return &Environment{store: s, consts: c, outer: nil}
}
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}
