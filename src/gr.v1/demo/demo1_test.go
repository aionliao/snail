package demo

import "testing"

func TestDemo1(t *testing.T) {
	d := &Demo1{}
	t.Log(d.myDo())
}

func TestDemo2(t *testing.T) {
	d := &Demo1{}
	t.Log(d.myDo())
}
