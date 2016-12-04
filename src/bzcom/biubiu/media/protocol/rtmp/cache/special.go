package cache

import (
	"bzcom/biubiu/media/av"
	"errors"
	"reflect"
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

func (self *SpecialCache)Full()bool{
	return self.full
}

func (self *SpecialCache)Equal(p av.Packet)bool{
	myLen := len(self.p.Data)
	inLen := len(p.Data)
	if myLen == inLen{
		return reflect.DeepEqual(p.Data, self.p.Data)
	}
	return false
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
