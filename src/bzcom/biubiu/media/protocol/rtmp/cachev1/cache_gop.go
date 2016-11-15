package cachev1

import "bzcom/biubiu/media/av"

type Gop struct {
}

func NewGop() *Gop {
	return &Gop{}
}

func (self *Gop) Write(p *av.Packet) {

}
