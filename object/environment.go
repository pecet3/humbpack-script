package object

type Environment struct {
	store   map[string]Object
	consts  map[string]Object
	modules map[string]Object
	public map[string]Object
	outer   *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok {
		obj, ok = e.consts[name]
	}
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
		if !ok {
			obj, ok = e.modules[name]
		}
	}
	return obj, ok
}

func (e *Environment) GetNoOuter(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok {
		obj, ok = e.consts[name]
	}
	return obj, ok
}

func (e *Environment) GetMutNoOuter(name string) (Object, bool) {
	obj, ok := e.store[name]
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
func (e *Environment) SetPublicConst(name string, val Object) Object {
	e.public[name] = val
	return val
}
func (e *Environment) Set(name string, obj Object) {
	if e.outer != nil {
		_, ok := e.outer.store[name]
		if ok {
			e.outer.store[name] = obj
		}
	}
	e.store[name] = obj
}
func (e *Environment) GetModule(name string) (Object, bool) {
	obj, ok := e.modules[name]
	return obj, ok
}

func (e *Environment) SetModule(name string, val Object) Object {
	e.modules[name] = val
	return val
}
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	c := make(map[string]Object)
	m := make(map[string]Object)
	return &Environment{store: s, consts: c, modules: m, outer: nil}
}
func NewClosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) WithOnlyPublic() {
	e.store = nil
	newConsts := make(map[string]Object)
	for k, v := range e.consts {
		if 
		newConsts[k] = v
	}
	e.consts = newConsts
}
