package cachev1

import "bzcom/biubiu/media/av"

type Metadata struct {
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func (self *Metadata) Write(p *av.Packet) {

}
