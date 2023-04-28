package framework

import (
	"context"
)

type Closer interface {
	Close(ctx context.Context)
}
type CloseFunc func(ctx context.Context)

func (f CloseFunc) Close(ctx context.Context) {
	f(ctx)
}

var defaultStack = NewStack()

func Add(name string, f Closer) { defaultStack.Add(name, f) }
func Close(ctx context.Context) { defaultStack.Close(ctx) }

type Function struct {
	Name   string
	Action Closer
}

type Stack struct {
	Functions []*Function
}

func NewStack() *Stack {
	return &Stack{
		Functions: make([]*Function, 0),
	}
}

func (s *Stack) Add(name string, f Closer) {
	s.push(&Function{
		Name:   name,
		Action: f,
	})
}

func (s *Stack) Close(ctx context.Context) {
	for i := s.pop(); i != nil; i = s.pop() {
		i.Action.Close(ctx)
	}
}

func (s *Stack) pop() *Function {
	if len(s.Functions) > 0 {
		v := s.Functions[len(s.Functions)-1]
		s.Functions = s.Functions[:len(s.Functions)-1]
		return v
	}
	return nil

}
func (s *Stack) push(f *Function) {
	s.Functions = append(s.Functions, f)
}
