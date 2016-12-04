package cache

import (
	"bzcom/biubiu/media/av"
	"errors"
	"time"
)

var (
	maxGOPCap   int = 1024
	ErrGopTooBig     = errors.New("gop to big")
)

type array struct {
	index   int
  timeID  int64
	packets []av.Packet
}

func newArray() array {
	ret := array{
		index:   0,
		timeID : time.Now().UnixNano()/1000000,
		packets: make([]av.Packet,0, maxGOPCap),
	}
	return ret
}

func (self *array)timeid()int64{
	return self.timeID
}

func (self *array) reset() {
	self.index = 0
	self.timeID = time.Now().UnixNano()/1000000
	self.packets = self.packets[:0]
}

func (self *array)end(index int)bool{
	return self.index == index
}

func (self *array) write(packet av.Packet) error {
	if self.index >= maxGOPCap {
		return ErrGopTooBig
	}
	self.packets = append(self.packets, packet)
	self.index = len(self.packets)-1
	return nil
}

func (self *array)read(index int)(packet av.Packet,nextindex int,err error){
	if index > self.index{
		return packet, nextindex, errors.New("invalid index")
	}else if index < self.index{
		nextindex = index+1
	}
	return self.packets[index],nextindex,nil
}


type GopCache struct {
	start     bool
	num       int
	count     int
	index     int
	gops      []array
}

func NewGopCache(num int) *GopCache {
	return &GopCache{
		count: num,
		gops:  make([]array, num),
	}
}

func (self *GopCache) writeToArray(p av.Packet, startNew bool) error {
	var ginc array
	if startNew{
		if self.num != self.count{
			ginc = newArray()
			self.num++
			self.index = self.num-1
		}else{
			self.index = (self.index + 1) % self.count
			ginc = self.gops[self.index]
			ginc.reset()
		}
	}else{
		ginc = self.gops[self.index]
	}
	ginc.write(p)
	self.gops[self.index] = ginc

	return nil
}

func (self *GopCache) Write(p av.Packet) {
	var ok bool
	if p.IsVideo {
		vh := p.Header.(av.VideoPacketHeader)
		if vh.IsKeyFrame() && !vh.IsSeq() {
			ok = true
		}
	}
	if ok || self.start {
		self.start = true
		self.writeToArray(p, ok)
	}
}

func (self *GopCache)Read(pos int, curid int64)(packet av.Packet,nextpos int,id int64,err error){
	var ginc array
	var index, nextindex,indexpos int

	if pos == -1{
		if self.num == self.count{
			index = (self.index+2)%self.count
		}
	}else{
			index = pos >> 16
			indexpos = pos & 0xffff
			ginc = self.gops[index]
			id := ginc.timeid()
			if id != curid &&  id > curid{
				indexpos = 0
				index = (index+2)%self.count
			}else if ginc.end(indexpos){
					indexpos = 0
					index = (index+1)%self.count
			}
	}
	ginc = self.gops[index]
	id = ginc.timeid()
	if id < curid{
		nextpos = pos
		id = curid
		err = Empty
		return
	}
	packet,nextindex,err = ginc.read(indexpos)
	nextpos = index<<16 | nextindex
	return
}
