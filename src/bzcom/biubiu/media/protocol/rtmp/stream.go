package rtmp

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/protocol/rtmp/cache"
	"errors"
	"sync"
	"time"
	"log"
)

var (
	EmptyID = ""
)

type RtmpStream struct {
	lock    sync.RWMutex
	streams map[string]*Stream
}

func NewRtmpStream() *RtmpStream {
	ret := &RtmpStream{
		streams: make(map[string]*Stream),
	}
	go ret.CheckAlive()
	return ret
}

func (self *RtmpStream) HandleReader(r av.ReadCloser) {
	self.lock.Lock()
	info := r.Info()
	s, ok := self.streams[info.Key]
	if !ok {
		s = NewStream()
		self.streams[info.Key] = s
	} else {
		s.ReaderStop()
		id := s.ID()
		if id != EmptyID && id != info.UID {
			ns := NewStream()
			s.Copy(ns)
			s = ns
			self.streams[info.Key] = ns
		}
	}
	s.AddReader(r)
	self.lock.Unlock()
}

func (self *RtmpStream) HandleWriter(w av.WriteCloser) {
	self.lock.Lock()
	info := w.Info()
	s, ok := self.streams[info.Key]
	if !ok {
		s = NewStream()
		self.streams[info.Key] = s
	}
	self.lock.Unlock()
	s.AddWriter(w)
}

func (self *RtmpStream) CheckAlive() {
	for {
		time.Sleep(time.Second * 5)
		self.lock.Lock()
		for k, v := range self.streams {
			if v.CheckAlive() == 0 {
				delete(self.streams, k)
			}
		}
		self.lock.Unlock()
	}
}

type Stream struct {
	lock    sync.RWMutex
	cache   *cache.Cache
	r       av.ReadCloser
	ws      map[string]mWriterCloser
}

type mWriterCloser struct {
	w    av.WriteCloser
}


func NewStream() *Stream {
	return &Stream{
		ws:    make(map[string]mWriterCloser),
	}
}

func (self *Stream) ID() string {
	if self.r != nil {
		return self.r.Info().UID
	}
	return EmptyID
}

func (self *Stream) Copy(dst *Stream) {
	for _, v := range self.ws {
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
	}
	dst.cache = self.cache
}

func (self *Stream) AddReader(r av.ReadCloser) {
	self.r = r
	if self.cache == nil{
		self.cache = cache.NewCache()
	}
	sc := newSourceConsole(r, *self.cache)
	go sc.do()
}

func (self *Stream) AddWriter(w av.WriteCloser) {
	info := w.Info()
	mw := mWriterCloser{
		w: w,
	}
	self.ws[info.UID] = mw
	if self.cache == nil{
		self.cache = cache.NewCache()
	}
	wc := newDestConsole(w, *self.cache)
	go wc.do()
}

func (self *Stream) ReaderStop() {
	if  self.r != nil {
		self.r.Close(errors.New("stop old"))
		for k, v := range self.ws {
			if v.w.Info().IsInterval() {
				v.w.Close(errors.New("close"))
				delete(self.ws, k)
			}
		}
	}
}

func (self *Stream) CheckAlive() (n int) {
	if self.r != nil  {
		if self.r.Alive() {
			n++
		} else {
			self.r.Close(errors.New("read timeout"))
		}
	}
	for k, v := range self.ws {
		if !v.w.Alive() {
			delete(self.ws, k)
			v.w.Close(errors.New("write timeout"))
			continue
		}
		n++
	}
	return
}

type sourceConsole struct{
	cache  cache.Cache
	r       av.ReadCloser
}

func newSourceConsole(r av.ReadCloser,c cache.Cache) sourceConsole{
	return sourceConsole{
		r:r,
		cache:c,
	}
}

func(self*sourceConsole)do(){
	var p av.Packet
	for {
		err := self.r.Read(&p)
		if err != nil {
			log.Println("s close")
			return
		}
		self.cache.Write(p)
	}
}


type destConsole struct{
	pos int
	flag int
	id int64
	c cache.Cache
	w av.WriteCloser
}


func newDestConsole( w av.WriteCloser,c cache.Cache) destConsole{
	return destConsole{
		w:w,
		c:c,
		pos:-1,
		flag:3,
	}
}

func (self *destConsole)do(){
	var err error
	var p av.Packet
	for{
		p, self.pos,self.id, err = self.c.Read(self.pos,self.flag,self.id)
		if err == cache.Empty{
			continue
		}
		if p.IsMetadata{
			self.flag=2
		}else{
			if !p.IsVideo {
				ah := p.Header.(av.AudioPacketHeader)
				if ah.SoundFormat() == av.SOUND_AAC &&
				   ah.AACPacketType() == av.AAC_SEQHDR {
					 self.flag=0
				}
			} else {
				vh := p.Header.(av.VideoPacketHeader)
				if vh.IsSeq() {
					self.flag=1
				}
			}
		}
		if err = self.w.Write(p);err !=nil{
			log.Println("w close")
			break
		}
	}
}
