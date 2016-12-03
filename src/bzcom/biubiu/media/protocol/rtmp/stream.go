package rtmp

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/protocol/rtmp/cache0"
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
		s.TransStop()
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
	isStart bool
	lock    sync.RWMutex
	cache   cache0.Cache
	r       av.ReadCloser
	ws      map[string]PackWriterCloser
}

type PackWriterCloser struct {
	init bool
	pos int
	flag int
	nextpos int
	id int64
	w    av.WriteCloser
}

func (self *PackWriterCloser)do(cache   cache0.Cache){
	var p av.Packet
	var err error
	self.pos=-1
	self.flag=3
	for{
		p, self.pos,self.id, err = cache.Read(self.pos,self.flag,self.id)
		if err == cache0.Empty{
			continue
		}
		log.Println("s",p.TimeStamp, p.IsVideo,p.IsMetadata,len(p.Data))
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
		self.w.Write(p)
	}
}

func NewStream() *Stream {
	return &Stream{
		cache: cache0.NewCache(),
		ws:    make(map[string]PackWriterCloser),
	}
}

func (self *Stream) ID() string {
	if self.r != nil {
		return self.r.Info().UID
	}
	return EmptyID
}

func (self *Stream) Copy(dst *Stream) {
	self.lock.Lock()
	for k, v := range self.ws {
		delete(self.ws, k)
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
	}
	self.lock.Unlock()
}

func (self *Stream) AddReader(r av.ReadCloser) {
	self.lock.Lock()
	self.r = r
	self.lock.Unlock()
	self.TransStart()
}

func (self *Stream) AddWriter(w av.WriteCloser) {
	self.lock.Lock()
	info := w.Info()
	pw := PackWriterCloser{
		w: w,
	}
	self.ws[info.UID] = pw
	go pw.do(self.cache)
	self.lock.Unlock()
}

func (self *Stream) TransStart() {
	go func() {
		self.isStart = true
		var p av.Packet
		for {
			if !self.isStart {
				self.closeInter()
				return
			}
			err := self.r.Read(&p)
			if err != nil {
				self.closeInter()
				self.isStart = false
				return
			}
			self.cache.Write(p)

			//self.lock.Lock()
			//
			// for k, v := range self.ws {
			// 	if !v.init {
			// 		if err = self.cache.Send(v.w); err != nil {
			// 			delete(self.ws, k)
			// 			continue
			// 		}
			// 		v.init = true
			// 		self.ws[k] = v
			// 	} else {
			// 		if err = v.w.Write(p); err != nil {
			// 			delete(self.ws, k)
			// 		}
			// 	}
			// }
			// self.lock.Unlock()
		}
	}()
}

func (self *Stream) TransStop() {
	if self.isStart && self.r != nil {
		self.r.Close(errors.New("stop old"))
	}
	self.isStart = false
}

func (self *Stream) CheckAlive() (n int) {
	self.lock.Lock()
	if self.r != nil && self.isStart {
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
	self.lock.Unlock()

	return
}

func (self *Stream) closeInter() {
	self.lock.Lock()
	for k, v := range self.ws {
		if v.w.Info().IsInterval() {
			v.w.Close(errors.New("close"))
			delete(self.ws, k)
		}
	}
	self.lock.Unlock()
}
