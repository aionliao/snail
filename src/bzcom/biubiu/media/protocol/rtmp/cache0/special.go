package cache0

import (
	"bytes"
	"bzcom/biubiu/media/av"
	"log"
	"errors"

	"bzcom/biubiu/media/protocol/amf"
)

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var (
	Empty = errors.New("empty")
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, amf.AMF0); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

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
