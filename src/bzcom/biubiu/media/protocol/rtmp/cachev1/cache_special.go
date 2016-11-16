package cachev1

import (
	"bytes"
	"bzcom/biubiu/media/av"
	"log"

	"bzcom/biubiu/media/protocol/amf"
)

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
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

type SpecialData struct {
	full bool
	p    av.Packet
}

func NewSpecialData() *SpecialData {
	return &SpecialData{}
}

func (self *SpecialData) Write(p *av.Packet) {
	if cap(self.p.Data) < len(p.Data) {
		self.p.Data = make([]byte, len(p.Data))
	}
	self.p.IsMetadata = p.IsMetadata
	self.p.IsVideo = p.IsVideo
	self.p.TimeStamp = p.TimeStamp
	copy(self.p.Data, p.Data)
	self.full = true
}

func (self *SpecialData) Send(w av.WriteCloser) error {
	if !self.full {
		return nil
	}
	if self.p.IsMetadata {
		self.p.Data = self.p.Data[len(setFrameFrame):]
	}
	return w.Write(self.p)
}
