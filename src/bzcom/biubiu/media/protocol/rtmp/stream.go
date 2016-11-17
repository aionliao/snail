package rtmp

import (
	"bzcom/biubiu/media/av"
	"bzcom/biubiu/media/protocol/rtmp/cachev1"
	"errors"
	"log"
	"sync"
	"time"
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
	s.AddWriter(w)
	self.lock.Unlock()
}

func (self *RtmpStream) CheckAlive() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			self.lock.Lock()
			for k, v := range self.streams {
				if v.CheckAlive() == 0 {
					delete(self.streams, k)
				}
			}
			self.lock.Unlock()
		default:
		}
	}
}

type Stream struct {
	isStart bool
	lock    sync.RWMutex
	cache   *cachev1.Cache
	r       av.ReadCloser
	ws      map[string]*PackWriterCloser
}

type PackWriterCloser struct {
	init bool
	w    av.WriteCloser
}

func NewStream() *Stream {
	return &Stream{
		cache: cachev1.NewCache(),
		ws:    make(map[string]*PackWriterCloser),
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
		v.w.Reset()
		dst.AddWriter(v.w)
	}
	self.lock.Unlock()
}

func (self *Stream) AddReader(r av.ReadCloser) {
	self.lock.Lock()
	log.Println("www", r)
	self.r = r
	self.TransStart()
	self.lock.Unlock()
}

func (self *Stream) AddWriter(w av.WriteCloser) {
	self.lock.Lock()
	log.Println("w", w)
	info := w.Info()
	pw := &PackWriterCloser{
		w: w,
	}
	self.ws[info.UID] = pw
	self.lock.Unlock()
}

func (self *Stream) TransStart() {
	go func() {
		self.isStart = true
		var p av.Packet
		for {
			if !self.isStart {
				return
			}
			err := self.r.Read(&p)
			if err != nil {
				self.isStart = false
				// TODO: close special writer
				return
			}

			self.cache.Write(&p)

			self.lock.Lock()
			for k, v := range self.ws {
				if !v.init {
					if err = self.cache.Send(v.w); err != nil {
						delete(self.ws, k)
						continue
					}
					v.init = true
				} else {
					if err = v.w.Write(p); err != nil {
						delete(self.ws, k)
					}
				}
			}
			self.lock.Unlock()
		}
	}()
}

func (self *Stream) TransStop() {
	if self.isStart && self.r != nil {
		self.r.Close(errors.New("stop old"))
		// TODO: close special writer
	}
	self.isStart = false
}

func (self *Stream) CheckAlive() (n int) {
	self.lock.Lock()
	if self.r != nil && self.isStart {
		if self.r.Alive() {
			n++
		}
	}
	for k, v := range self.ws {
		if !v.w.Alive() {
			delete(self.ws, k)
			continue
		}
		n++
	}
	self.lock.Unlock()

	return
}
