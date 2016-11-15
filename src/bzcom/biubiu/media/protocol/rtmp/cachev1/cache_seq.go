package cachev1

import "bzcom/biubiu/media/av"

type Seq struct {
}

func NewSeq() *Seq {
	return &Seq{}
}

func (self *Seq) Write(p *av.Packet) {

}
