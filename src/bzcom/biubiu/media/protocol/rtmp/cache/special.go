package cache

import (
	"bzcom/biubiu/media/av"
	"errors"
)

var (
	Empty = errors.New("empty")
)

type SpecialCache struct {
	full bool
	p    av.Packet
}

func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

func (self *SpecialCache) Write(p av.Packet) {
	self.p = p
	self.full = true
}


func (self *SpecialCache)Read()(av.Packet,error){
	if self.full{
		return self.p, nil
	}
	return self.p,Empty
}
