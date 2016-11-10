package demo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDemo1(t *testing.T) {
	at := assert.New(t)
	d := &Demo1{}
	at.Equal(d.myDo(), 1)
}
