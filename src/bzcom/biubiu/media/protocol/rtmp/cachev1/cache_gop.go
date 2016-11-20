package cachev1

import (
	"bzcom/biubiu/media/av"
	"errors"
)

type checker struct {
	start    bool
	getVideo bool
	curTs    uint32
	firstTs  uint32
	interval uint32
}

func newChecker() *checker {
	return &checker{
		interval: 5 * 1000,
	}
}

func (self *checker) do(p *av.Packet) bool {
	ok := false
	self.curTs = p.TimeStamp

	if !self.start {
		ok = true
		self.start = true
	} else {
		if p.IsVideo {
			self.getVideo = true
			vh := p.Header.(av.VideoPacketHeader)
			if vh.IsKeyFrame() && !vh.IsSeq() {
				ok = true
			}
		}
		v := self.curTs - self.firstTs
		if !self.getVideo && v >= self.interval {
			ok = true
		}
	}
	if ok {
		self.getVideo = false
		self.firstTs = self.curTs
	}

	return ok
}

var (
	maxGOPSzie   int = 1024
	ErrGopTooBig     = errors.New("gop to big")
)

type array struct {
	index   int
	packets []av.Packet
}

func newArray() *array {
	ret := &array{
		index:   0,
		packets: make([]av.Packet, maxGOPSzie),
	}
	return ret
}

func (self *array) reset() {
	self.index = 0
	self.packets = self.packets[:0]
}

func (self *array) write(packet av.Packet) error {
	if self.index >= maxGOPSzie {
		return ErrGopTooBig
	}
	self.packets = append(self.packets, packet)
	self.index++
	return nil
}

func (self *array) send(w av.WriteCloser) error {
	var err error
	for i := 0; i < self.index; i++ {
		packet := self.packets[i]
		if err = w.Write(packet); err != nil {
			return err
		}
	}
	return err
}

type GopCache struct {
	start     bool
	num       int
	count     int
	nextindex int
	gops      []*array
	check     *checker
}

func NewGopCache(num int) *GopCache {
	return &GopCache{
		count: num,
		check: newChecker(),
		gops:  make([]*array, num),
	}
}

func (self *GopCache) writeToArray(chunk av.Packet, startNew bool) error {
	var ginc *array
	if startNew {
		ginc = self.gops[self.nextindex]
		if ginc == nil {
			ginc = newArray()
			self.num++
			self.gops[self.nextindex] = ginc
		} else {
			ginc.reset()
		}
		self.nextindex = (self.nextindex + 1) % self.count
	} else {
		ginc = self.gops[(self.nextindex+1)%self.count]
	}
	ginc.write(chunk)

	return nil
}

func (self *GopCache) Write(p av.Packet) {
	ok := self.check.do(&p)
	if ok || self.start {
		self.start = true
		self.writeToArray(p, ok)
	}
}

func (self *GopCache) sendTo(w av.WriteCloser) error {
	var err error
	pos := (self.nextindex + 1) % self.count
	for i := 0; i < self.num; i++ {
		index := (pos - self.num + 1) + i
		if index < 0 {
			index += self.count
		}
		g := self.gops[index]
		err = g.send(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *GopCache) Send(w av.WriteCloser) error {
	return self.sendTo(w)
}
