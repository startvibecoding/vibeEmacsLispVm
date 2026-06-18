package vm

import "sort"

// Env is a lexical environment.
type Env struct {
	parent *Env
	values map[string]Value
}

// NewEnv creates an empty environment with an optional parent.
func NewEnv(parent *Env) *Env {
	return &Env{parent: parent, values: make(map[string]Value)}
}

// Get returns a value bound to name.
func (e *Env) Get(name string) (Value, bool) {
	if e == nil {
		return nil, false
	}
	if v, ok := e.values[name]; ok {
		return v, true
	}
	return e.parent.Get(name)
}

// Define binds name in the current environment.
func (e *Env) Define(name string, v Value) {
	e.values[name] = v
}

// Set updates an existing binding, or defines it in the current environment.
func (e *Env) Set(name string, v Value) {
	if e == nil {
		return
	}
	if _, ok := e.values[name]; ok {
		e.values[name] = v
		return
	}
	if e.parent != nil {
		if _, ok := e.parent.Get(name); ok {
			e.parent.Set(name, v)
			return
		}
	}
	e.values[name] = v
}

// Names returns names bound in this environment and its parents in sorted order.
func (e *Env) Names() []string {
	seen := make(map[string]struct{})
	for cur := e; cur != nil; cur = cur.parent {
		for name := range cur.values {
			seen[name] = struct{}{}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
