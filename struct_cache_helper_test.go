package cache

import (
	"testing"
)

func TestHelperPointer(t *testing.T) {
	type A struct{}

	a := A{}
	b := &a
	c := "hello"
	d := &c
	e := make(map[string]string)
	f := &e
	g := []string{}
	h := &g
	i := 11
	j := &i
	k := 11.11
	l := &k

	if isPointer(a) {
		t.Errorf("expected a is not poiter, got - it is")
	}
	if !isPointer(b) {
		t.Errorf("expected a is poiter, got - it is not")
	}
	if isPointer(c) {
		t.Errorf("expected a is not poiter, got - it is")
	}
	if !isPointer(d) {
		t.Errorf("expected a is poiter, got - it is not")
	}
	if isPointer(e) {
		t.Errorf("expected a is not poiter, got - it is")
	}
	if !isPointer(f) {
		t.Errorf("expected a is poiter, got - it is not")
	}
	if isPointer(g) {
		t.Errorf("expected a is not poiter, got - it is")
	}
	if !isPointer(h) {
		t.Errorf("expected a is poiter, got - it is not")
	}
	if isPointer(i) {
		t.Errorf("expected a is not poiter, got - it is")
	}
	if !isPointer(j) {
		t.Errorf("expected a is poiter, got - it is not")
	}
	if isPointer(k) {
		t.Errorf("expected a is not poiter, got - it is")
	}
	if !isPointer(l) {
		t.Errorf("expected a is poiter, got - it is not")
	}
}
