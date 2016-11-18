package cachev1

import "bzcom/biubiu/media/av"

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
	return w.Write(self.p)
}
