package vm

import (
	"fmt"

	"github.com/pecet3/hmbk-script/object"
)

// VMObject reprezentuje obiekt w naszej VM
type VMObject struct {
	Name     string
	Value    object.Object
	Marked   bool
	Children []*VMObject // referencje do innych obiektów
}

// VM przechowuje Stack i Heap VM
type VM struct {
	Heap      []*VMObject
	Stack     []*VMObject // obiekty w użyciu
	IntReg    [32]int64
	StrReg    [32]string
	FloatReg  [32]float64
	ObjectReg [32]*VMObject
	PC        int
}

// Dodanie obiektu do Heapu VM
func (vm *VM) Alloc(name string) *VMObject {
	obj := &VMObject{Name: name}
	vm.Heap = append(vm.Heap, obj)
	return obj
}

// Mark: oznacz obiekty osiągalne ze Stacka
func (vm *VM) mark() {
	var markObj func(obj *VMObject)
	markObj = func(obj *VMObject) {
		if obj == nil || obj.Marked {
			return
		}
		obj.Marked = true
		for _, child := range obj.Children {
			markObj(child)
		}
	}

	for _, root := range vm.Stack {
		markObj(root)
	}
}

// Sweep: usuwa nieoznaczone obiekty z Heapu
func (vm *VM) sweep() {
	var newHeap []*VMObject
	for _, obj := range vm.Heap {
		if obj.Marked {
			obj.Marked = false // reset marker na przyszły cykl
			newHeap = append(newHeap, obj)
		} else {
			fmt.Println("Collecting:", obj.Name)
		}
	}
	vm.Heap = newHeap
}
func (obj *VMObject) AddChildren(children ...*VMObject) {
	obj.Children = append(obj.Children, children...)
}

// Usuwa podane dzieci z obiektu
func (obj *VMObject) RemoveChildren(children ...*VMObject) {
	newChildren := obj.Children[:0] // tworzymy nowy slice, zachowując pojemność
	for _, c := range obj.Children {
		remove := false
		for _, r := range children {
			if c == r {
				remove = true
				break
			}
		}
		if !remove {
			newChildren = append(newChildren, c)
		}
	}
	obj.Children = newChildren
}
func (vm *VM) GC() {
	vm.mark()
	vm.sweep()
}

func Run() {
	vm := &VM{}

	// Tworzymy obiekty
	a := vm.Alloc("A")
	b := vm.Alloc("B")
	c := vm.Alloc("C")
	d := vm.Alloc("D")

	// Tworzymy referencje
	a.AddChildren(b, c)
	b.AddChildren(d)

	// Stack wskazuje na obiekt A
	vm.Stack = []*VMObject{a}

	fmt.Println("Heap przed GC:", len(vm.Heap)) // 4
	vm.GC()                                     // Mark & Sweep
	fmt.Println("Heap po GC:", len(vm.Heap))    // 4, wszystkie osiągalne

	// Usuwamy referencję do B
	a.Children = []*VMObject{c}
	fmt.Println("Usuwamy referencję do B")
	vm.GC()                                  // GC powinien zebrać B i D
	fmt.Println("Heap po GC:", len(vm.Heap)) // 2
}
