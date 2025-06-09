package utils

import (
	"errors"
)

type Stack[T any] struct{ elements []T }

// Returns the element at the top of the stack without removing it
func NewStack[T any](init ...T) Stack[T] {
	if len(init) == 0 {
		return Stack[T]{}
	}

	return Stack[T]{elements: init}
}

// Returns the element at the top of the stack without removing it
func (stack *Stack[T]) Top() (T, error) {
	if stack.Count() == 0 {
		var zero T
		return zero, errors.New("unexpected stack size of 0, cannot Top()")
	}

	index := len(stack.elements) - 1
	top := stack.elements[index]
	return top, nil
}

// Returns the count of the element in the stack (stack-size)
func (stack *Stack[T]) Count() int {
	return len(stack.elements)
}

// Push a new 'element' onto the stack
func (stack *Stack[T]) Push(elem T) {
	stack.elements = append(stack.elements, elem)
}

// Removes an element from the stack top, if the stack size is
// already 0 (stack empty) it returns an error
func (stack *Stack[T]) Pop() (T, error) {
	if stack.Count() == 0 {
		var zero T // The only way to instantiate a generic type zero value
		return zero, errors.New("unexpected stack size of 0, cannot Pop()")
	}

	index := len(stack.elements) - 1
	top := stack.elements[index]
	stack.elements = stack.elements[:index]
	return top, nil
}

func (stack *Stack[T]) Iterator() func(yield func(T) bool) {
	return func(yield func(T) bool) {
		for i := len(stack.elements) - 1; i >= 0; i-- {
			if !yield(stack.elements[i]) {
				return
			}
		}
	}
}
