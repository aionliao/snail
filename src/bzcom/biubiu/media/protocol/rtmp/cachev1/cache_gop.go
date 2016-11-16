package cachev1

import (
	"bzcom/biubiu/media/av"
	"time"
)

const (
	bits24MaxValue int64 = 0xffffff
	bits32MaxValue int64 = 0xffffffff
)

type signed struct {
	curTs         int64
	firstTs       int64
	curAt         time.Time
	startAt       time.Time
	begin         bool
	hasVideo      bool
	hasSetFirstTs bool
	interval      time.Duration
	maxDuration   time.Duration
}

func newSigned() *signed {
	return &signed{
		interval:    5 * time.Second,
		maxDuration: 6 * time.Second,
	}
}

func (self *signed) do(p *av.Packet) bool {
	sign := false
	nowTime := time.Now()
	self.curAt = nowTime
	self.curTs = int64(p.TimeStamp)

	if !self.begin {
		sign = true
		self.begin = true
	} else {
		flag := false
		if p.IsVideo {
			vh := p.Header.(av.VideoPacketHeader)
			if vh.IsKeyFrame() && !vh.IsSeq() {
				flag = true
			}
		}
		interval := self.getTimeDiffInterval()
		if (!self.hasVideo && interval >= int64(self.interval/time.Millisecond)) ||
			(self.hasVideo && interval >= int64(self.maxDuration/time.Millisecond)) {
			flag = true
		}
		if flag {
			sign = true
		}
	}

	if sign {
		self.hasVideo = false
		self.startAt = nowTime
		self.firstTs = int64(p.TimeStamp)
	} else {
		if p.IsVideo {
			self.hasVideo = true
		}
	}

	return sign
}

func (self *signed) getTimeDiffInterval() int64 {
	diff := self.curTs - self.firstTs
	if diff < 0 {
		tmpdiff := int64(0)
		if self.firstTs <= bits24MaxValue {
			tmpdiff = bits24MaxValue + diff + 1
		} else if self.firstTs > bits24MaxValue &&
			self.firstTs <= bits32MaxValue {
			tmpdiff = bits32MaxValue + diff + 1
		}
		return tmpdiff
	}
	return diff
}

//每一个GOP会持有N个msg
type gop struct {
	index  int
	chunks []av.Packet
}

var (
	gopLen int = 100
)

func newGop() gop {
	ret := gop{
		index:  0,
		chunks: make([]av.Packet, gopLen),
	}
	return ret
}

func (g *gop) pushback(chunk av.Packet) error {
	if g.index == len(g.chunks) {
		newchunks := make([]av.Packet, gopLen)
		g.chunks = append(g.chunks, newchunks...)
	}

	g.chunks[g.index] = chunk
	g.index++
	return nil
}

func (g *gop) send(w av.WriteCloser) error {
	var err error
	for i := 0; i < g.index; i++ {
		cc := g.chunks[i]
		err = w.Write(cc)
		if err != nil {
			return err
		}
	}
	return err
}

func (g *gop) clean() {
	g.index = 0
	g.chunks = g.chunks[:0]
}

type gops struct {
	index int
	num   int
	count int
	gops  []gop
}

func newGops(count int) *gops {
	return &gops{
		count: count,
		gops:  make([]gop, count),
	}
}

func (q *gops) pushback(chunk av.Packet, sign bool) error {
	var g gop
	if sign {
		if q.num < q.count {
			g = newGop()
			q.index = q.num
			q.num++
		} else {
			q.index = (q.index + 1) % q.count
			g = q.gops[q.index]
			g.clean()
		}
	} else {
		g = q.gops[q.index]
	}
	g.pushback(chunk)
	q.gops[q.index] = g

	return nil
}

func (self *gops) send(w av.WriteCloser) error {
	var err error
	for i := 0; i < self.num; i++ {
		index := (self.index - self.num + 1) + i
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

type Gop struct {
	mark bool
	q    *gops
	s    *signed
}

func NewGop(num int) *Gop {
	return &Gop{
		q: newGops(num),
		s: newSigned(),
	}
}

func (self *Gop) Write(p *av.Packet) {
	var ap av.Packet
	ap = *p
	ap.Data = make([]byte, len(p.Data))
	copy(ap.Data, p.Data)

	sign := self.s.do(&ap)
	if self.mark == false && sign || self.mark {
		self.mark = true
		self.q.pushback(ap, sign)
	}
}

func (self *Gop) Send(w av.WriteCloser) error {
	return self.q.send(w)
}
